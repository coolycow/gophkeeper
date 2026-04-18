package grpcserver

import (
	"context"
	"strings"

	"github.com/coolycow/gophkeeper/internal/logger"
	"github.com/coolycow/gophkeeper/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// metadataSetAuthorization — нестандартный исходящий заголовок с hex-токеном для следующих вызовов (аналог Set-Cookie).
const metadataSetAuthorization = "set-authorization"

// AuthUnaryServerInterceptor подставляет userID в контекст по правилам OptionalAuth (HTTP).
func AuthUnaryServerInterceptor(users service.UserService) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Получаем metadata из контекста
		md, ok := metadata.FromIncomingContext(ctx)

		// Если metadata нет, создаём пустой metadata
		if !ok {
			md = metadata.MD{}
		}

		// Получаем authorization из metadata
		auth := ""
		if v := md.Get("authorization"); len(v) > 0 {
			auth = v[0]
		}

		// Решаем пользователя и токен
		userID, newToken, err := resolveGRPCUser(ctx, users, auth)
		if err != nil {
			return nil, err
		}

		// Если токен изменился, устанавливаем его в metadata
		if newToken != "" {
			if herr := grpc.SetHeader(ctx, metadata.Pairs(metadataSetAuthorization, newToken)); herr != nil {
				logger.Log.Error("grpc auth: set header", zap.Error(herr))
				return nil, status.Error(codes.Internal, "internal error")
			}
		}

		// Кладём userID в контекст
		ctx = shortener.ContextWithUserID(ctx, userID)
		return handler(ctx, req)
	}
}

// resolveGRPCUser: пустой/невалидный токен → новый пользователь и новый токен; валидный, но user не найден → Unauthenticated.
func resolveGRPCUser(ctx context.Context, users service.UserService, authHeader string) (userID string, newAuthHex string, err error) {
	authHeader = strings.TrimSpace(authHeader)

	userID, decErr := users.GetUserIDFromAuthToken(authHeader)
	if decErr != nil {
		// Если токен невалидный, создаём нового пользователя
		u, cerr := users.CreateUser(ctx)
		if cerr != nil {
			logger.Log.Error("grpc auth: create user", zap.Error(cerr))
			return "", "", status.Error(codes.Internal, "internal error")
		}

		// Получаем токен для нового пользователя
		tok, terr := users.GetCookieValueByUser(u)
		if terr != nil {
			logger.Log.Error("grpc auth: issue token", zap.Error(terr))
			return "", "", status.Error(codes.Internal, "internal error")
		}

		return u.ID, tok, nil
	}

	// Если пользователь не найден, возвращаем ошибку Unauthenticated
	if _, gerr := users.GetUser(ctx, userID); gerr != nil {
		return "", "", status.Error(codes.Unauthenticated, "user not found")
	}

	return userID, "", nil
}
