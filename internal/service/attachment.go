package service

import (
	"context"

	"github.com/coolycow/gophkeeper/internal/config"
	"github.com/coolycow/gophkeeper/internal/model"
	"github.com/coolycow/gophkeeper/internal/repository"
	"github.com/go-playground/validator/v10"
)

// AttachmentService Сервис для работы с вложениями
type AttachmentService interface {
	GetAttachmentByID(ctx context.Context, userID string, attachmentID string) (*model.Attachment, error)                                 // получает вложение по его ID
	GetAttachmentsBySecretID(ctx context.Context, userID string, secretID string) ([]*model.Attachment, error)                            // получает все вложения по ID секрета
	GetAttachmentsBySecretVersionID(ctx context.Context, userID string, secretVersionID string) ([]*model.Attachment, error)              // получает все вложения по ID версии секрета
	CreateAttachment(ctx context.Context, userID string, secretVersionID string, attachment *model.Attachment) (*model.Attachment, error) // создает вложение
	HardDeleteAttachment(ctx context.Context, userID string, attachmentID string) error                                                   // удаляет вложение
}

// Реализация сервисного слоя
type attachmentService struct {
	repo      repository.GophKeeperRepository
	cfg       *config.ConfigServer
	validator *validator.Validate
}

// NewAttachmentService инициализация сервиса
func NewAttachmentService(cfg *config.ConfigServer, repo repository.GophKeeperRepository) AttachmentService {
	return &attachmentService{
		repo:      repo,
		cfg:       cfg,
		validator: validator.New(),
	}
}

// GetAttachmentByID получает вложение по его ID
func (s *attachmentService) GetAttachmentByID(ctx context.Context, userID string, attachmentID string) (*model.Attachment, error) {
	return s.repo.GetAttachmentByID(ctx, userID, attachmentID)
}

// GetAttachmentsBySecretID получает все вложения по ID секрета
func (s *attachmentService) GetAttachmentsBySecretID(ctx context.Context, userID string, secretID string) ([]*model.Attachment, error) {
	return s.repo.GetAttachmentsBySecretID(ctx, userID, secretID)
}

// GetAttachmentsBySecretVersionID получает все вложения по ID версии секрета
func (s *attachmentService) GetAttachmentsBySecretVersionID(ctx context.Context, userID string, secretVersionID string) ([]*model.Attachment, error) {
	return s.repo.GetAttachmentsBySecretVersionID(ctx, userID, secretVersionID)
}

// CreateAttachment создает вложение
func (s *attachmentService) CreateAttachment(ctx context.Context, userID string, secretVersionID string, attachment *model.Attachment) (*model.Attachment, error) {
	return s.repo.CreateAttachment(ctx, userID, secretVersionID, attachment)
}

// HardDeleteAttachment удаляет вложение
func (s *attachmentService) HardDeleteAttachment(ctx context.Context, userID string, attachmentID string) error {
	return s.repo.HardDeleteAttachment(ctx, userID, attachmentID)
}
