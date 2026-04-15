package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"

	keeperError "github.com/coolycow/gophkeeper/internal/error"

	"github.com/coolycow/gophkeeper/internal/config"
	"github.com/coolycow/gophkeeper/internal/model"
	"github.com/coolycow/gophkeeper/internal/repository"
)

// UserService Сервис для работы с пользователями
type UserService interface {
	// Методы получения пользователя
	GetUserByID(ctx context.Context, userID string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByEmailAndPassword(ctx context.Context, email string, password string) (*model.User, error)

	// Методы для работы с пользователем
	CreateUser(ctx context.Context, request model.UserRegisterRequest) (*model.User, error)
	UpdateUser(ctx context.Context, userID string, request model.UserUpdateRequest) error

	// Методы для работы с мягким и полным удалением пользователя
	SoftDeleteUser(ctx context.Context, userID string) error
	HardDeleteUser(ctx context.Context, userID string) error

	// Методы для работы с куками
	GetUserIDFromCookie(cookie *http.Cookie) (string, error)
	GetCookieValueByUser(user model.User) (string, error)
	GetCookieValueByUserID(userID string) (string, error)

	// GetUserIDFromAuthToken — то же значение, что и cookie auth (metadata authorization / Bearer).
	GetUserIDFromAuthToken(token string) (string, error)
}

// Реализация сервисного слоя
type userService struct {
	repo      repository.GophKeeperRepository
	cfg       *config.ConfigServer
	validator *validator.Validate
}

// NewUserService инициализация сервиса
func NewUserService(cfg *config.ConfigServer, repo repository.GophKeeperRepository) UserService {
	return &userService{
		repo:      repo,
		cfg:       cfg,
		validator: validator.New(),
	}
}

// GetUserByID возвращает пользователя по его ID
func (s *userService) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

// GetUserByEmail возвращает пользователя по его email
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	// Валидация email
	if err := s.validator.Var(email, "required,email"); err != nil {
		return nil, keeperError.CustomError{
			Message:    fmt.Sprintf("email validation failed: %s", err),
			StatusCode: http.StatusBadRequest,
		}
	}

	// Получаем пользователя по email
	return s.repo.GetUserByEmail(ctx, email)
}

// GetUserByEmailAndPassword возвращает пользователя по его email и паролю
func (s *userService) GetUserByEmailAndPassword(ctx context.Context, email string, password string) (*model.User, error) {
	// Получаем пользователя по email (email будет валидирован в GetUserByEmail)
	user, err := s.GetUserByEmail(ctx, email)

	// Если ошибка, возвращаем ошибку (уже тип keeperError.CustomError)
	if err != nil {
		return nil, err
	}

	// Если пользователь не найден, возвращаем ошибку
	if user == nil {
		return nil, keeperError.CustomError{
			Message:    "user not found",
			StatusCode: http.StatusUnauthorized,
		}
	}

	// Сравниваем пароль с хешем в базе данных
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	// Если пароль не совпадает, возвращаем ошибку
	if err != nil {
		return nil, keeperError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusUnauthorized,
		}
	}

	return user, nil
}

// CreateUser создаёт нового пользователя
func (s *userService) CreateUser(ctx context.Context, request model.UserRegisterRequest) (*model.User, error) {
	// Валидация почты
	if err := s.validator.Var(request.Email, "required,email"); err != nil {
		return nil, keeperError.CustomError{
			Message:    fmt.Sprintf("email validation failed: %s", err),
			StatusCode: http.StatusBadRequest,
		}
	}

	// Валидация пароля
	if err := validatePassword(request.Password, s.cfg, s.validator); err != nil {
		return nil, keeperError.CustomError{
			Message:    fmt.Sprintf("password validation failed: %s", err),
			StatusCode: http.StatusBadRequest,
		}
	}

	// Проверяем, существует ли пользователь с такой email
	existingUser, err := s.repo.GetUserByEmail(ctx, request.Email)

	if err != nil {
		return nil, keeperError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	// Если пользователь с такой email уже существует, возвращаем ошибку
	if existingUser != nil {
		return nil, keeperError.CustomError{
			Message:    "user with this email already exists",
			StatusCode: http.StatusConflict,
		}
	}

	// Хешируем пароль
	hashedPassword, err := hashPassword(request.Password)

	// Если ошибка хеширования пароля, возвращаем ошибку
	if err != nil {
		return nil, keeperError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	// Генерируем соль
	salt, err := generateSalt(s.cfg.SaltLength)

	// Если ошибка генерации соли, возвращаем ошибку
	if err != nil {
		return nil, keeperError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	// Создаём пользователя
	return s.repo.CreateUser(ctx, &model.User{
		Email:     request.Email,
		Password:  hashedPassword,
		Salt:      hex.EncodeToString(salt),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
}

// UpdateUser обновляет пользователя
func (s *userService) UpdateUser(ctx context.Context, userID string, request model.UserUpdateRequest) error {
	// Проверяем, существует ли пользователь с таким ID
	user, err := s.repo.GetUserByID(ctx, userID)

	if err != nil {
		return keeperError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	if user == nil {
		return keeperError.CustomError{
			Message:    "user not found",
			StatusCode: http.StatusNotFound,
		}
	}

	// Обновляем пароль если он изменился
	if request.OldPassword != "" && request.NewPassword != "" {
		// Сравниваем старый пароль с хешем в базе данных, если он не совпадает, возвращаем ошибку
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.OldPassword))
		if err != nil {
			return keeperError.CustomError{
				Message:    "old password is incorrect",
				StatusCode: http.StatusUnauthorized,
			}
		}

		// Валидация нового пароля
		if err := validatePassword(request.NewPassword, s.cfg, s.validator); err != nil {
			return keeperError.CustomError{
				Message:    fmt.Sprintf("password validation failed: %s", err),
				StatusCode: http.StatusBadRequest,
			}
		}

		// Хешируем новый пароль
		hashedPassword, err := hashPassword(request.NewPassword)
		if err != nil {
			return keeperError.CustomError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			}
		}

		// Обновляем пароль
		user.Password = hashedPassword
	}

	// Валидация email если email изменился
	if request.Email != "" && request.Email != user.Email {
		if err := s.validator.Var(request.Email, "email"); err != nil {
			return keeperError.CustomError{
				Message:    fmt.Sprintf("email validation failed: %s", err),
				StatusCode: http.StatusBadRequest,
			}
		}

		// Проверяем, существует ли пользователь с такой email
		existingUser, err := s.repo.GetUserByEmail(ctx, request.Email)

		if err != nil {
			return keeperError.CustomError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			}
		}

		// Если пользователь с такой email уже существует, возвращаем ошибку
		if existingUser != nil {
			return keeperError.CustomError{
				Message:    "user with this email already exists",
				StatusCode: http.StatusConflict,
			}
		}

		// Обновляем email
		user.Email = request.Email
	}

	// TODO: Если изменился пароль, то все данные должны быть зашифрованы заново

	// Обновляем пользователя
	return s.repo.UpdateUser(ctx, userID, &model.User{
		Email:     request.Email, // Email может быть изменен или не изменен
		Password:  user.Password, // Пароль может быть изменен или не изменен
		Salt:      user.Salt,     // Соль не изменяется
		UpdatedAt: time.Now(),
	})
}

// SoftDeleteUser мягко удаляет пользователя
func (s *userService) SoftDeleteUser(ctx context.Context, userID string) error {
	return s.repo.SoftDeleteUser(ctx, userID)
}

// HardDeleteUser полностью удаляет пользователя
func (s *userService) HardDeleteUser(ctx context.Context, userID string) error {
	return s.repo.HardDeleteUser(ctx, userID)
}

// GetUserIDFromCookie достаёт UserID из переданной куки
func (s *userService) GetUserIDFromCookie(cookie *http.Cookie) (string, error) {
	return s.GetUserIDFromAuthToken(cookie.Value)
}

// GetUserIDFromAuthToken декодирует hex-токен (gRPC metadata authorization).
func (s *userService) GetUserIDFromAuthToken(token string) (string, error) {
	// Убираем пробелы и префикс "bearer "
	token = strings.TrimSpace(token)
	if len(token) > 6 && strings.EqualFold(token[:7], "bearer ") {
		token = strings.TrimSpace(token[7:])
	}

	// Сохраняем значение токена
	cookieValue := token

	// Декодируем hex
	data, err := hex.DecodeString(cookieValue)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex cookie value: %w", err)
	}

	if len(data) == 0 {
		return "", errors.New("invalid cookie")
	}

	key := sha256.Sum256([]byte(s.cfg.SecretKey))

	aesBlock, err := aes.NewCipher(key[:])
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	// создаём вектор инициализации
	nonceSize := aesGCM.NonceSize()

	// Разделяем nonce и зашифрованные данные
	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	// расшифровываем
	decrypted, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

// GetCookieValueByUser возвращает значение куки для указанного user
func (s *userService) GetCookieValueByUser(user model.User) (string, error) {
	return s.GetCookieValueByUserID(user.ID)
}

// GetCookieValueByUserID возвращает значение куки для указанного userID
func (s *userService) GetCookieValueByUserID(userID string) (string, error) {
	key := sha256.Sum256([]byte(s.cfg.SecretKey))

	aesBlock, err := aes.NewCipher(key[:])
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	// создаём вектор инициализации
	nonce, err := generateRandom(aesGCM.NonceSize())
	if err != nil {
		return "", fmt.Errorf("failed to generate random nonce: %w", err)
	}

	dst := aesGCM.Seal(nil, nonce, []byte(userID), nil)

	// Сохраняем nonce вместе с зашифрованными данными
	result := append(nonce, dst...)

	return hex.EncodeToString(result), nil
}
