package grpcserver

import (
	"context"

	"github.com/coolycow/gophkeeper/internal/ctxutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// requireUserID возвращает userID из контекста (после auth-interceptor) или Unauthenticated.
func requireUserID(ctx context.Context) (string, error) {
	id, ok := ctxutil.UserIDFromContext(ctx)
	if !ok || id == "" {
		return "", status.Error(codes.Unauthenticated, "missing user context")
	}
	return id, nil
}
