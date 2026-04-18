// Package ctxutil хранит в context.Context идентификатор пользователя для gRPC/HTTP-обработчиков.
package ctxutil

import "context"

type userIDKey struct{}

// WithUserID возвращает контекст с привязанным userID (для вызовов после успешной авторизации).
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// UserIDFromContext достаёт userID, если он был установлен (например в gRPC-interceptor).
func UserIDFromContext(ctx context.Context) (string, bool) {
	// Пользователь может быть аутентифицирован в HTTP- или gRPC-контексте.
	v := ctx.Value(userIDKey{})

	// Если пользователь не аутентифицирован, возвращаем false.
	if v == nil {
		return "", false
	}

	// Попробуем получить userID из контекста.
	s, ok := v.(string)

	// Если userID не является строкой или пустой, возвращаем false.
	if !ok || s == "" {
		return "", false
	}

	// Если пользователь аутентифицирован, возвращаем его userID.
	return s, true
}
