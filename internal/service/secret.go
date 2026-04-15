package service

import (
	"context"
	"net/http"
	"time"

	keeperError "github.com/coolycow/gophkeeper/internal/error"

	"github.com/coolycow/gophkeeper/internal/config"
	"github.com/coolycow/gophkeeper/internal/model"
	"github.com/coolycow/gophkeeper/internal/repository"
	"github.com/go-playground/validator/v10"
)

// SecretService Сервис для работы с секретами
type SecretService interface {
	// Методы для получения секрета
	GetSecretByID(ctx context.Context, secretID string) (*model.Secret, error)
	GetSecretsByUserID(ctx context.Context, userID string) ([]*model.Secret, error)
	GetSecretByUserIDAndSecretID(ctx context.Context, userID string, secretID string) (*model.Secret, error)

	// Методы для работы с секретами
	CreateSecret(ctx context.Context, userID string, request model.SecretCreateRequest) (*model.Secret, error)
	UpdateSecret(ctx context.Context, userID string, secretID string, request model.SecretUpdateRequest) error
	SoftDeleteSecret(ctx context.Context, userID string, secretID string) error
	HardDeleteSecret(ctx context.Context, userID string, secretID string) error

	// Специальные методы для работы с секретами
	CompressSecretByID(ctx context.Context, userID string, secretID string) error
	CompressSecretsByUserID(ctx context.Context, userID string) error
}

// Реализация сервисного слоя
type secretService struct {
	repo      repository.GophKeeperRepository
	cfg       *config.ConfigServer
	validator *validator.Validate
}

// NewSecretService инициализация сервиса
func NewSecretService(cfg *config.ConfigServer, repo repository.GophKeeperRepository) SecretService {
	return &secretService{
		repo:      repo,
		cfg:       cfg,
		validator: validator.New(),
	}
}

// GetSecretByID получает секрет по его ID
func (s *secretService) GetSecretByID(ctx context.Context, secretID string) (*model.Secret, error) {
	return s.repo.GetSecretByID(ctx, secretID)
}

// GetSecretsByUserID получает все секреты пользователя
func (s *secretService) GetSecretsByUserID(ctx context.Context, userID string) ([]*model.Secret, error) {
	return s.repo.GetSecretsByUserID(ctx, userID)
}

// GetSecretByUserIDAndSecretID получает секрет по его ID и ID пользователя
func (s *secretService) GetSecretByUserIDAndSecretID(ctx context.Context, userID string, secretID string) (*model.Secret, error) {
	return s.repo.GetSecretByUserIDAndSecretID(ctx, userID, secretID)
}

// CreateSecret создает секрет
func (s *secretService) CreateSecret(ctx context.Context, userID string, request model.SecretCreateRequest) (*model.Secret, error) {
	// Создаём объект секрета
	// ID текущей версии секрета генерируется автоматически в репозитории в момент создания секрета
	secret := &model.Secret{
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: time.Time{},
		SecretVersions: []*model.SecretVersion{
			{
				DataEncrypted: request.DataEncrypted,
			},
		},
	}

	return s.repo.CreateSecret(ctx, userID, secret)
}

// UpdateSecret обновляет секрет
func (s *secretService) UpdateSecret(ctx context.Context, userID string, secretID string, request model.SecretUpdateRequest) error {
	// Получаем секрет
	secret, err := s.repo.GetSecretByUserIDAndSecretID(ctx, userID, secretID)

	// Если произошла ошибка, возвращаем её
	if err != nil {
		return keeperError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	// Если секрет не найден, возвращаем ошибку
	if secret == nil {
		return keeperError.CustomError{
			Message:    "secret not found",
			StatusCode: http.StatusNotFound,
		}
	}

	// Обновляем секрет
	secret.SecretVersions = append(secret.SecretVersions, &model.SecretVersion{
		DataEncrypted: request.DataEncrypted,
	})

	return s.repo.UpdateSecret(ctx, userID, secretID, secret)
}

// SoftDeleteSecret мягко удаляет секрет
func (s *secretService) SoftDeleteSecret(ctx context.Context, userID string, secretID string) error {
	return s.repo.SoftDeleteSecret(ctx, userID, secretID)
}

// HardDeleteSecret полностью удаляет секрет
func (s *secretService) HardDeleteSecret(ctx context.Context, userID string, secretID string) error {
	return s.repo.HardDeleteSecret(ctx, userID, secretID)
}

// CompressSecretByID удаляет все версии секрета кроме последней
// Восстановление удалённых версий секрета невозможно
func (s *secretService) CompressSecretByID(ctx context.Context, userID string, secretID string) error {
	return s.repo.CompressSecretByID(ctx, userID, secretID)
}

// CompressSecretsByUserID удаляет все версии секретов кроме последних для всех секретов пользователя
// Восстановление удалённых версий секретов невозможно
func (s *secretService) CompressSecretsByUserID(ctx context.Context, userID string) error {
	return s.repo.CompressSecretsByUserID(ctx, userID)
}
