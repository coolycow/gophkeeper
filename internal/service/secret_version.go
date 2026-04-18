package service

import (
	"context"

	"net/http"

	"github.com/coolycow/gophkeeper/internal/config"
	keeperError "github.com/coolycow/gophkeeper/internal/error"
	"github.com/coolycow/gophkeeper/internal/model"
	"github.com/coolycow/gophkeeper/internal/repository"
	"github.com/go-playground/validator/v10"
)

// SecretVersionService Сервис для работы с версиями секретов
type SecretVersionService interface {
	// Методы для получения версии секрета
	GetLatestSecretVersion(ctx context.Context, userID string, secretID string) (*model.SecretVersion, error)
	GetSecretVersionByID(ctx context.Context, userID string, secretID string, secretVersionID string) (*model.SecretVersion, error)
	GetAllSecretHistories(ctx context.Context, userID string, secretID string) ([]*model.SecretVersion, error)
	CreateSecretVersion(ctx context.Context, userID string, secretID string, secretVersion *model.SecretVersion) (*model.SecretVersion, error)
	HardDeleteSecretVersion(ctx context.Context, userID string, secretID string, secretVersionID string) error
	RestoreSecretVersion(ctx context.Context, userID string, secretID string, secretVersionID string) error
}

// Реализация сервисного слоя
type secretVersionService struct {
	repo      repository.GophKeeperRepository
	cfg       *config.ConfigServer
	validator *validator.Validate
}

// NewSecretVersionService инициализация сервиса
func NewSecretVersionService(cfg *config.ConfigServer, repo repository.GophKeeperRepository) SecretVersionService {
	return &secretVersionService{
		repo:      repo,
		cfg:       cfg,
		validator: validator.New(),
	}
}

// GetLatestSecretVersion получает последнюю версию секрета
func (s *secretVersionService) GetLatestSecretVersion(ctx context.Context, userID string, secretID string) (*model.SecretVersion, error) {
	return s.repo.GetLatestSecretVersion(ctx, userID, secretID)
}

// GetSecretVersionByID получает версию секрета по его ID
func (s *secretVersionService) GetSecretVersionByID(ctx context.Context, userID string, secretID string, secretVersionID string) (*model.SecretVersion, error) {
	return s.repo.GetSecretVersionByID(ctx, userID, secretID, secretVersionID)
}

// GetAllSecretHistories получает все версии секрета
func (s *secretVersionService) GetAllSecretHistories(ctx context.Context, userID string, secretID string) ([]*model.SecretVersion, error) {
	return s.repo.GetAllSecretHistories(ctx, userID, secretID)
}

// CreateSecretVersion создает версию секрета
func (s *secretVersionService) CreateSecretVersion(ctx context.Context, userID string, secretID string, secretVersion *model.SecretVersion) (*model.SecretVersion, error) {
	// Проверяем, что секрет не существует и принадлежит пользователю
	secret, err := s.repo.GetSecretByUserIDAndSecretID(ctx, userID, secretID)

	// Ошибка получения секрета
	if err != nil {
		return nil, err
	}

	// Если секрет не найден, возвращаем ошибку
	if secret == nil {
		return nil, keeperError.CustomError{
			Message:    "secret not found",
			StatusCode: http.StatusNotFound,
		}
	}

	// Проверяем общее количество версий секрета
	secretVersions, err := s.repo.GetAllSecretHistories(ctx, userID, secretID)

	if err != nil {
		return nil, err
	}

	// Если количество версий секрета равно максимальному количеству версий, то удаляем самую старую версию.
	// Особый случай: если у секрета всего одна версия, то удаляем её и по сути секрет становится пустым.
	// Если s.cfg.SecretVersionCount == 0, то не ограничиваем количество версий.
	if s.cfg.SecretVersionCount > 0 && len(secretVersions) >= s.cfg.SecretVersionCount {
		err = s.repo.HardDeleteOldestSecretVersion(ctx, userID, secretID)
		if err != nil {
			return nil, err
		}
	}

	// Вычисляем размер данных
	secretVersion.DataSize = len(secretVersion.DataEncrypted)

	return s.repo.CreateSecretVersion(ctx, userID, secretID, secretVersion)
}

// HardDeleteSecretVersion полностью удаляет версию секрета (мягкое удаление невозможно, т.к. не имеет смысла)
func (s *secretVersionService) HardDeleteSecretVersion(ctx context.Context, userID string, secretID string, secretVersionID string) error {
	return s.repo.HardDeleteSecretVersion(ctx, userID, secretID, secretVersionID)
}

// RestoreSecretVersion восстанавливает версию секрета (делает её текущей для секрета)
func (s *secretVersionService) RestoreSecretVersion(ctx context.Context, userID string, secretID string, secretVersionID string) error {
	return s.repo.RestoreSecretVersion(ctx, userID, secretID, secretVersionID)
}
