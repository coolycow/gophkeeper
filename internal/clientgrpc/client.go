package clientgrpc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/coolycow/gophkeeper/internal/proto/gophkeeperpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Client обертка для GophKeeperServiceClient с обработкой токенов и одной автоматической подстановкой при Unauthenticated.
type Client struct {
	mu sync.Mutex

	svc  gophkeeperpb.GophKeeperServiceClient
	conn *grpc.ClientConn

	accessToken  string
	refreshToken string
	saltHex      string
	expiresAt    time.Time
}

// NewClient создает API клиент над существующим gRPC соединением.
func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		svc:  gophkeeperpb.NewGophKeeperServiceClient(conn),
		conn: conn,
	}
}

// Conn возвращает основное соединение (для Close из main).
func (c *Client) Conn() *grpc.ClientConn { return c.conn }

// Close закрывает gRPC соединение.
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// ApplyAuthResponse сохраняет токены из Register, Login или RefreshToken.
func (c *Client) ApplyAuthResponse(resp *gophkeeperpb.AuthResponse) error {
	if resp == nil {
		return errors.New("clientgrpc: nil auth response")
	}

	// Блокируем доступ к токенам для синхронизации
	c.mu.Lock()
	defer c.mu.Unlock()

	// Сохраняем токены
	c.accessToken = strings.TrimSpace(resp.GetToken())
	c.refreshToken = strings.TrimSpace(resp.GetRefreshToken())
	c.saltHex = strings.TrimSpace(resp.GetSalt())

	// Проверяем, что токены не пусты
	if c.accessToken == "" || c.refreshToken == "" {
		return errors.New("clientgrpc: missing token in auth response")
	}

	// Сохраняем время истечения доступа
	if ts := resp.GetExpiresAt(); ts != nil {
		c.expiresAt = ts.AsTime()
	} else {
		c.expiresAt = time.Time{}
	}

	return nil
}

// SaltHex возвращает соль пользователя из последнего успешного auth (hex строка для secretcrypto).
func (c *Client) SaltHex() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.saltHex
}

// AccessExpiresAt возвращает время истечения доступа сервером (если есть).
func (c *Client) AccessExpiresAt() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.expiresAt
}

// RefreshToken возвращает текущий refresh токен для сохранения сессии.
func (c *Client) RefreshToken() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.refreshToken
}

// Register вызывает публичный RPC Register.
func (c *Client) Register(ctx context.Context, email, password string) error {
	// Вызываем публичный RPC Register
	resp, err := c.svc.Register(ctx, &gophkeeperpb.RegisterRequest{Email: email, Password: password})

	if err != nil {
		return err
	}

	return c.ApplyAuthResponse(resp)
}

// Login вызывает публичный RPC Login.
func (c *Client) Login(ctx context.Context, email, password string) error {
	// Вызываем публичный RPC Login
	resp, err := c.svc.Login(ctx, &gophkeeperpb.LoginRequest{Email: email, Password: password})

	if err != nil {
		return err
	}

	return c.ApplyAuthResponse(resp)
}

// Refresh обменивает refresh токен на новую пару (без Bearer заголовка).
func (c *Client) Refresh(ctx context.Context) error {
	// Блокируем доступ к refresh токену для синхронизации
	c.mu.Lock()
	rt := c.refreshToken
	c.mu.Unlock()

	// Проверяем, что refresh токен не пустой
	if rt == "" {
		return errors.New("clientgrpc: no refresh token")
	}

	// Вызываем RPC RefreshToken
	resp, err := c.svc.RefreshToken(ctx, &gophkeeperpb.RefreshTokenRequest{RefreshToken: rt})

	if err != nil {
		return err
	}

	return c.ApplyAuthResponse(resp)
}

// Ping проверяет доступность сервера.
func (c *Client) Ping(ctx context.Context) (string, error) {
	// Вызываем RPC Ping
	resp, err := c.svc.Ping(ctx, &gophkeeperpb.PingRequest{})

	if err != nil {
		return "", err
	}

	return resp.GetStatus(), nil
}

// bearer возвращает строку с префиксом Bearer и токеном
func bearer(token string) string {
	return "Bearer " + strings.TrimSpace(token)
}

// withAccess добавляет заголовок authorization с токеном в контекст
func withAccess(ctx context.Context, access string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", bearer(access))
}

// isUnauthenticated проверяет, является ли ошибка Unauthenticated.
func isUnauthenticated(err error) bool {
	if err == nil {
		return false
	}

	st, ok := status.FromError(err)

	return ok && st.Code() == codes.Unauthenticated
}

// call выполняет аутентифицированный unary RPC; при Unauthenticated выполняет Refresh и повторяет.
func (c *Client) call(ctx context.Context, fn func(ctx context.Context) error) error {
	// Блокируем доступ к access токену для синхронизации
	c.mu.Lock()
	access := c.accessToken
	c.mu.Unlock()

	// Проверяем, что access токен не пустой
	if access == "" {
		return errors.New("clientgrpc: not authenticated")
	}

	// Вызываем функцию с токеном
	err := fn(withAccess(ctx, access))

	// Проверяем, является ли ошибка Unauthenticated
	if isUnauthenticated(err) {
		// Обновляем токены с помощью refresh токена
		if rerr := c.Refresh(ctx); rerr != nil {
			return fmt.Errorf("refresh after 401: %w", rerr)
		}
		c.mu.Lock()
		access = c.accessToken
		c.mu.Unlock()
		err = fn(withAccess(ctx, access))
	}
	return err
}

// ListSecrets возвращает краткое описание секретов для данного scope.
func (c *Client) ListSecrets(ctx context.Context, scope gophkeeperpb.SecretListScope) (*gophkeeperpb.ListSecretsResponse, error) {
	var out *gophkeeperpb.ListSecretsResponse
	err := c.call(ctx, func(ctx context.Context) error {
		resp, err := c.svc.ListSecrets(ctx, &gophkeeperpb.ListSecretsRequest{ListScope: scope})
		if err != nil {
			return err
		}
		out = resp
		return nil
	})
	return out, err
}

// GetSecret загружает один секрет (при необходимости с историей версий).
func (c *Client) GetSecret(ctx context.Context, secretID string, withHistory bool) (*gophkeeperpb.Secret, error) {
	var out *gophkeeperpb.Secret
	err := c.call(ctx, func(ctx context.Context) error {
		resp, err := c.svc.GetSecret(ctx, &gophkeeperpb.GetSecretRequest{
			SecretId:              secretID,
			IncludeVersionHistory: withHistory,
		})
		if err != nil {
			return err
		}
		out = resp
		return nil
	})
	return out, err
}

// CreateSecret отправляет первую зашифрованную версию.
func (c *Client) CreateSecret(ctx context.Context, dataEncrypted, titleEncrypted []byte, dataFormatVersion int32) (*gophkeeperpb.Secret, error) {
	var out *gophkeeperpb.Secret
	err := c.call(ctx, func(ctx context.Context) error {
		resp, err := c.svc.CreateSecret(ctx, &gophkeeperpb.CreateSecretRequest{
			DataEncrypted:     dataEncrypted,
			TitleEncrypted:    titleEncrypted,
			DataFormatVersion: dataFormatVersion,
		})
		if err != nil {
			return err
		}
		out = resp
		return nil
	})
	return out, err
}

// UpdateSecret отправляет новую зашифрованную версию.
func (c *Client) UpdateSecret(ctx context.Context, secretID string, dataEncrypted, titleEncrypted []byte, dataFormatVersion int32) (*gophkeeperpb.SecretVersion, error) {
	var out *gophkeeperpb.SecretVersion
	err := c.call(ctx, func(ctx context.Context) error {
		resp, err := c.svc.UpdateSecret(ctx, &gophkeeperpb.UpdateSecretRequest{
			SecretId:          secretID,
			DataEncrypted:     dataEncrypted,
			TitleEncrypted:    titleEncrypted,
			DataFormatVersion: dataFormatVersion,
		})
		if err != nil {
			return err
		}
		out = resp
		return nil
	})
	return out, err
}

// DeleteSecret мягко удаляет секрет на сервере.
func (c *Client) DeleteSecret(ctx context.Context, secretID string) error {
	return c.call(ctx, func(ctx context.Context) error {
		_, err := c.svc.DeleteSecret(ctx, &gophkeeperpb.DeleteSecretRequest{SecretId: secretID})
		return err
	})
}

// ListSecretVersions возвращает все версии указанного секрета.
func (c *Client) ListSecretVersions(ctx context.Context, secretID string) ([]*gophkeeperpb.SecretVersion, error) {
	var out []*gophkeeperpb.SecretVersion
	err := c.call(ctx, func(ctx context.Context) error {
		resp, err := c.svc.ListSecretVersions(ctx, &gophkeeperpb.ListSecretVersionsRequest{SecretId: secretID})
		if err != nil {
			return err
		}
		out = resp.GetVersions()
		return nil
	})
	return out, err
}

// RestoreSecretVersion делает выбранную версию текущей для секрета.
func (c *Client) RestoreSecretVersion(ctx context.Context, secretID, secretVersionID string) error {
	return c.call(ctx, func(ctx context.Context) error {
		_, err := c.svc.RestoreSecretVersion(ctx, &gophkeeperpb.RestoreSecretVersionRequest{
			SecretId:        secretID,
			SecretVersionId: secretVersionID,
		})
		return err
	})
}

// DeleteSecretVersion удаляет версию секрета.
func (c *Client) DeleteSecretVersion(ctx context.Context, secretID, secretVersionID string) error {
	return c.call(ctx, func(ctx context.Context) error {
		_, err := c.svc.DeleteSecretVersion(ctx, &gophkeeperpb.DeleteSecretVersionRequest{
			SecretId:        secretID,
			SecretVersionId: secretVersionID,
		})
		return err
	})
}

// NeedsRefreshSoon возвращает true, если access токен истекает в пределах margin (для проактивного обновления).
func NeedsRefreshSoon(expiresAt time.Time, margin time.Duration, now time.Time) bool {
	if expiresAt.IsZero() {
		return false
	}
	return !now.Add(margin).Before(expiresAt)
}
