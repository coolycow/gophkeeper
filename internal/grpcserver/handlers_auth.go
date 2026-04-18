package grpcserver

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/coolycow/gophkeeper/internal/model"
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
	user, err := s.userSvc.CreateUser(ctx, model.UserRegisterRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return nil, grpcError(err)
	}

	return s.newAuthResponse(user)
}

// Login выполняет вход по email и паролю.
func (s *Server) Login(ctx context.Context, req *gophkeeperpb.LoginRequest) (*gophkeeperpb.AuthResponse, error) {
	user, err := s.userSvc.GetUserByEmailAndPassword(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, grpcError(err)
	}

	return s.newAuthResponse(user)
}

// RefreshToken выдаёт новую пару токенов по ранее выданному refresh-токену (тот же формат hex, что и access).
func (s *Server) RefreshToken(ctx context.Context, req *gophkeeperpb.RefreshTokenRequest) (*gophkeeperpb.AuthResponse, error) {
	rt := strings.TrimSpace(req.GetRefreshToken())
	if rt == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token required")
	}

	userID, err := s.userSvc.GetUserIDFromAuthToken(rt)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	u, err := s.userSvc.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.Unauthenticated, "user not found")
		}
		return nil, grpcError(err)
	}
	if u == nil {
		return nil, status.Error(codes.Unauthenticated, "user not found")
	}
	if !u.DeletedAt.IsZero() {
		return nil, status.Error(codes.Unauthenticated, "user deleted")
	}

	return s.newAuthResponse(u)
}

// newAuthResponse создаёт AuthResponse с токенами и сроком действия.
func (s *Server) newAuthResponse(u *model.User) (*gophkeeperpb.AuthResponse, error) {
	tok, err := s.userSvc.GetCookieValueByUser(*u)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "issue token: %v", err)
	}

	exp := timestamppb.New(time.Now().Add(sessionTokenTTL))
	return &gophkeeperpb.AuthResponse{
		Salt:           u.Salt,
		Token:          tok,
		RefreshToken:   tok,
		ExpiresAt:      exp,
	}, nil
}
