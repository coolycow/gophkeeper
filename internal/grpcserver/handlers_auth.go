package grpcserver

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/coolycow/gophkeeper/internal/model"
	"github.com/coolycow/gophkeeper/internal/observer/audit"
	"github.com/coolycow/gophkeeper/internal/proto/gophkeeperpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// sessionTokenTTL задаёт срок жизни, отражаемый в AuthResponse.expires_at (сами hex-токены пока без встроенного expiry).
const sessionTokenTTL = 24 * time.Hour

// Ping проверяет доступность gRPC-сервиса.
func (s *Server) Ping(_ context.Context, _ *gophkeeperpb.PingRequest) (*gophkeeperpb.PingResponse, error) {
	return &gophkeeperpb.PingResponse{Status: "ok"}, nil
}

// Register создаёт пользователя и возвращает токены доступа.
func (s *Server) Register(ctx context.Context, req *gophkeeperpb.RegisterRequest) (*gophkeeperpb.AuthResponse, error) {
	// Пытаемся создать пользователя с переданными email и паролем.
	user, err := s.userSvc.CreateUser(ctx, model.UserRegisterRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})

	// Если ошибка, возвращаем ошибку.
	if err != nil {
		return nil, grpcError(err)
	}

	// Если пользователь успешно создан, возвращаем токены доступа.
	resp, err := s.newAuthResponse(user)
	if err != nil {
		return nil, err
	}
	s.emitAudit(audit.ActionRegister, user.ID, "", "", "")
	return resp, nil
}

// Login выполняет вход по email и паролю.
func (s *Server) Login(ctx context.Context, req *gophkeeperpb.LoginRequest) (*gophkeeperpb.AuthResponse, error) {
	// Пытаемся получить пользователя по email и паролю.
	user, err := s.userSvc.GetUserByEmailAndPassword(ctx, req.GetEmail(), req.GetPassword())

	// Если ошибка, возвращаем ошибку.
	if err != nil {
		return nil, grpcError(err)
	}

	// Если пользователь успешно найден, возвращаем токены доступа.
	resp, err := s.newAuthResponse(user)
	if err != nil {
		return nil, err
	}

	// Отправляем событие аудита
	s.emitAudit(audit.ActionLogin, user.ID, "", "", "")

	// Возвращаем токены доступа
	return resp, nil
}

// RefreshToken выдаёт новую пару токенов по ранее выданному refresh-токену (тот же формат hex, что и access).
func (s *Server) RefreshToken(ctx context.Context, req *gophkeeperpb.RefreshTokenRequest) (*gophkeeperpb.AuthResponse, error) {
	// Пытаемся получить refresh-токен из запроса.
	rt := strings.TrimSpace(req.GetRefreshToken())
	if rt == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token required")
	}

	// Пытаемся получить userID из refresh-токена.
	userID, err := s.userSvc.GetUserIDFromAuthToken(rt)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	// Пытаемся получить пользователя по userID.
	u, err := s.userSvc.GetUserByID(ctx, userID)

	// Если ошибка, возвращаем ошибку.
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.Unauthenticated, "user not found")
		}
		return nil, grpcError(err)
	}

	// Если пользователь не найден, возвращаем ошибку.
	if u == nil {
		return nil, status.Error(codes.Unauthenticated, "user not found")
	}

	// Если пользователь удален (мягкое удаление), возвращаем ошибку.
	if u.DeletedAt != nil && !u.DeletedAt.IsZero() {
		return nil, status.Error(codes.Unauthenticated, "user deleted")
	}

	// Если пользователь успешно найден, возвращаем токены доступа.
	resp, err := s.newAuthResponse(u)
	if err != nil {
		return nil, err
	}

	// Отправляем событие аудита
	s.emitAudit(audit.ActionRefreshToken, u.ID, "", "", "")

	// Возвращаем токены доступа
	return resp, nil
}

// newAuthResponse создаёт AuthResponse с токенами и сроком действия токенов.
func (s *Server) newAuthResponse(u *model.User) (*gophkeeperpb.AuthResponse, error) {
	// Пытаемся получить токены доступа и refresh-токен для пользователя.
	tok, err := s.userSvc.GetCookieValueByUser(*u)

	// Если ошибка, возвращаем ошибку.
	if err != nil {
		return nil, status.Errorf(codes.Internal, "issue token: %v", err)
	}

	// Создаём срок действия токенов.
	exp := timestamppb.New(time.Now().Add(sessionTokenTTL))

	// Возвращаем токены доступа и refresh-токен.
	return &gophkeeperpb.AuthResponse{
		Salt:         u.Salt,
		Token:        tok,
		RefreshToken: tok,
		ExpiresAt:    exp,
	}, nil
}
