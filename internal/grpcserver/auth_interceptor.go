package grpcserver

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/coolycow/gophkeeper/internal/ctxutil"
	"github.com/coolycow/gophkeeper/internal/logger"
	"github.com/coolycow/gophkeeper/internal/proto/gophkeeperpb"
	"github.com/coolycow/gophkeeper/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthUnaryServerInterceptor проверяет metadata "authorization" для защищённых методов
// и кладёт userID в контекст. Публичные методы (см. isPublicGRPCMethod) пропускаются без токена.
func AuthUnaryServerInterceptor(users service.UserService) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if isPublicGRPCMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}

		auth := ""
		if v := md.Get("authorization"); len(v) > 0 {
			auth = strings.TrimSpace(v[0])
		}

		userID, err := resolveProtectedUser(ctx, users, auth)
		if err != nil {
			return nil, err
		}

		ctx = ctxutil.WithUserID(ctx, userID)
		return handler(ctx, req)
	}
}

// isPublicGRPCMethod: методы, которые не требуют авторизации (выполняются без токена)
func isPublicGRPCMethod(fullMethod string) bool {
	switch fullMethod {
	case gophkeeperpb.GophKeeperService_Ping_FullMethodName,
		gophkeeperpb.GophKeeperService_Register_FullMethodName,
		gophkeeperpb.GophKeeperService_Login_FullMethodName,
		gophkeeperpb.GophKeeperService_RefreshToken_FullMethodName:
		return true
	default:
		return false
	}
}

// resolveProtectedUser: для защищённых RPC нужен валидный токен и существующий пользователь.
func resolveProtectedUser(ctx context.Context, users service.UserService, authHeader string) (userID string, err error) {
	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" {
		return "", status.Error(codes.Unauthenticated, "missing authorization")
	}

	userID, decErr := users.GetUserIDFromAuthToken(authHeader)
	if decErr != nil {
		return "", status.Error(codes.Unauthenticated, "invalid or expired token")
	}

	u, gerr := users.GetUserByID(ctx, userID)
	if gerr != nil {
		if errors.Is(gerr, sql.ErrNoRows) {
			return "", status.Error(codes.Unauthenticated, "user not found")
		}
		logger.Log.Error("grpc auth: get user", zap.Error(gerr))
		return "", status.Error(codes.Internal, "internal error")
	}
	if u == nil {
		return "", status.Error(codes.Unauthenticated, "user not found")
	}

	return userID, nil
}
