package grpcserver

import (
	"context"
	"strings"

	"github.com/coolycow/gophkeeper/internal/model"
	"github.com/coolycow/gophkeeper/internal/observer/audit"
	"github.com/coolycow/gophkeeper/internal/proto/gophkeeperpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Ping проверяет доступность gRPC-сервиса.
func (s *Server) Ping(_ context.Context, _ *gophkeeperpb.PingRequest) (*gophkeeperpb.PingResponse, error) {
	return &gophkeeperpb.PingResponse{Status: "ok"}, nil
}

// Register создаёт пользователя и возвращает токены доступа.
func (s *Server) Register(ctx context.Context, req *gophkeeperpb.RegisterRequest) (*gophkeeperpb.AuthResponse, error) {
	// Создаём пользователя
	user, err := s.userSvc.CreateUser(ctx, model.UserRegisterRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})

	// Если ошибка при создании пользователя, возвращаем ошибку
	if err != nil {
		return nil, grpcError(err)
	}

	resp, err := s.newAuthResponse(ctx, user)
	if err != nil {
		return nil, err
	}

	// Отправляем событие аудита
	s.emitAudit(audit.ActionRegister, user.ID, "", "", "")

	// Возвращаем токены доступа
	return resp, nil
}

// Login выполняет вход по email и паролю.
func (s *Server) Login(ctx context.Context, req *gophkeeperpb.LoginRequest) (*gophkeeperpb.AuthResponse, error) {
	// Получаем пользователя по email и паролю
	user, err := s.userSvc.GetUserByEmailAndPassword(ctx, req.GetEmail(), req.GetPassword())

	// Если ошибка при получении пользователя, возвращаем ошибку
	if err != nil {
		return nil, grpcError(err)
	}

	// Получаем токены доступа
	resp, err := s.newAuthResponse(ctx, user)
	if err != nil {
		return nil, err
	}

	// Отправляем событие аудита
	s.emitAudit(audit.ActionLogin, user.ID, "", "", "")

	// Возвращаем токены доступа
	return resp, nil
}

// RefreshToken выдаёт новую пару токенов по ранее выданному refresh-токену (хранится в БД по хэшу).
func (s *Server) RefreshToken(ctx context.Context, req *gophkeeperpb.RefreshTokenRequest) (*gophkeeperpb.AuthResponse, error) {
	// Очищаем refresh-токен от пробелов
	rt := strings.TrimSpace(req.GetRefreshToken())

	// Если refresh-токен пустой, возвращаем ошибку
	if rt == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token required")
	}

	// Обмениваем refresh-токен на новые токены
	u, access, refresh, exp, err := s.userSvc.ExchangeRefreshToken(ctx, rt)
	if err != nil {
		return nil, grpcError(err)
	}

	// Отправляем событие аудита
	s.emitAudit(audit.ActionRefreshToken, u.ID, "", "", "")

	// Возвращаем токены доступа
	return &gophkeeperpb.AuthResponse{
		Salt:         u.Salt,
		Token:        access,
		RefreshToken: refresh,
		ExpiresAt:    timestamppb.New(exp),
	}, nil
}

// newAuthResponse создаёт AuthResponse: access — JWT, refresh — непрозрачная строка (в БД только хэш).
func (s *Server) newAuthResponse(ctx context.Context, u *model.User) (*gophkeeperpb.AuthResponse, error) {
	// Выдаём новые токены доступа
	access, refresh, exp, err := s.userSvc.IssueAuthTokens(ctx, u)

	// Если ошибка при выдаче токенов, возвращаем ошибку
	if err != nil {
		return nil, status.Errorf(codes.Internal, "issue token: %v", err)
	}

	// Возвращаем токены доступа
	return &gophkeeperpb.AuthResponse{
		Salt:         u.Salt,
		Token:        access,
		RefreshToken: refresh,
		ExpiresAt:    timestamppb.New(exp),
	}, nil
}
