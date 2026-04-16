// Package gophkeeper — общая логика сокращения ссылок для HTTP (Gin) и gRPC.
package gophkeeper

import "context"

// ctxUserIDKey — внутренний ключ для userID в context.Context (после auth).
type ctxUserIDKey struct{}

// ContextWithUserID кладёт идентификатор пользователя в контекст (используется gRPC-интерцептором).
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ctxUserIDKey{}, userID)
}

// UserIDFromContext читает userID, положенный ContextWithUserID.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(ctxUserIDKey{})
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok && s != ""
}
