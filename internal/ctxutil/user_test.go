package ctxutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Тестируем методы для работы с идентификатором пользователя в контексте
func TestWithUserID_UserIDFromContext_roundTrip(t *testing.T) {
	const want = "user-42"
	ctx := WithUserID(context.Background(), want)
	got, ok := UserIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, want, got)
}

// Нет пользователя в контексте
func TestUserIDFromContext_noUser(t *testing.T) {
	id, ok := UserIDFromContext(context.Background())
	assert.False(t, ok)
	assert.Empty(t, id)
}

// Пустая строка в контексте
func TestUserIDFromContext_emptyStringStored(t *testing.T) {
	ctx := WithUserID(context.Background(), "")
	id, ok := UserIDFromContext(ctx)
	assert.False(t, ok)
	assert.Empty(t, id)
}

// Неправильный тип данных в контексте (ожидается строка, а не число)
func TestUserIDFromContext_wrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), userIDKey{}, 123)
	id, ok := UserIDFromContext(ctx)
	assert.False(t, ok)
	assert.Empty(t, id)
}
