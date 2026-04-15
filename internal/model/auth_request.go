package model

// LoginRequest модель авторизации
// Валидация в сервисе user.go
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest модель регистрации
// Валидация в сервисе user.go
type UserRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UpdateRequest модель обновления пользователя
// Если пользователь изменил пароль, то все данные должны быть зашифрованы заново на клиенте и переданы на сервер в этом запросе
type UserUpdateRequest struct {
	Email       string                 `json:"email,omitempty"`
	OldPassword string                 `json:"old_password,omitempty"`
	NewPassword string                 `json:"new_password,omitempty"`
	Secrets     []*SecretCreateRequest `json:"secrets,omitempty"`
}
