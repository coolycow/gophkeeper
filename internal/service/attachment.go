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
	GetAttachmentsBySecretID(ctx context.Context, userID string, secretID string) ([]*model.AttachmentSummary, error)                     // получает все вложения по ID секрета (без тела data_encrypted)
	GetAttachmentsBySecretVersionID(ctx context.Context, userID string, secretVersionID string) ([]*model.AttachmentSummary, error)       // получает все вложения по ID версии секрета (без тела data_encrypted)
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
func (s *attachmentService) GetAttachmentsBySecretID(ctx context.Context, userID string, secretID string) ([]*model.AttachmentSummary, error) {
	return s.repo.GetAttachmentsBySecretID(ctx, userID, secretID)
}

// GetAttachmentsBySecretVersionID получает все вложения по ID версии секрета
func (s *attachmentService) GetAttachmentsBySecretVersionID(ctx context.Context, userID string, secretVersionID string) ([]*model.AttachmentSummary, error) {
	return s.repo.GetAttachmentsBySecretVersionID(ctx, userID, secretVersionID)
}

// CreateAttachment создает вложение
func (s *attachmentService) CreateAttachment(ctx context.Context, userID string, secretVersionID string, attachment *model.Attachment) (*model.Attachment, error) {
	attachment.InfoSize = len(attachment.InfoEncrypted)
	attachment.DataSize = len(attachment.DataEncrypted)

	if attachment.DataFormatVersion == 0 {
		attachment.DataFormatVersion = 1
	}

	if attachment.InfoFormatVersion == 0 {
		attachment.InfoFormatVersion = 1
	}

	return s.repo.CreateAttachment(ctx, userID, secretVersionID, attachment)
}

// HardDeleteAttachment удаляет вложение
func (s *attachmentService) HardDeleteAttachment(ctx context.Context, userID string, attachmentID string) error {
	return s.repo.HardDeleteAttachment(ctx, userID, attachmentID)
}
