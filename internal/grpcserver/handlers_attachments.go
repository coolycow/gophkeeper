package grpcserver

import (
	"context"

	"github.com/coolycow/gophkeeper/internal/model"
	"github.com/coolycow/gophkeeper/internal/observer/audit"
	"github.com/coolycow/gophkeeper/internal/proto/gophkeeperpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListAttachments возвращает вложения по secret_id или secret_version_id (ровно одно поле oneof).
func (s *Server) ListAttachments(ctx context.Context, req *gophkeeperpb.ListAttachmentsRequest) (*gophkeeperpb.ListAttachmentsResponse, error) {
	// Проверяем, что пользователь авторизован
	userID, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}

	// Проверяем, что secret_id или secret_version_id не пустые
	sid := req.GetSecretId()
	svid := req.GetSecretVersionId()

	// Проверяем, что secret_id или secret_version_id не пустые
	if sid != "" && svid != "" {
		return nil, status.Error(codes.InvalidArgument, "specify either secret_id or secret_version_id")
	}

	// Проверяем, что secret_id или secret_version_id не пустые
	if sid == "" && svid == "" {
		return nil, status.Error(codes.InvalidArgument, "secret_id or secret_version_id required")
	}

	// Получаем все вложения по secret_id или secret_version_id
	var list []*model.AttachmentSummary
	if svid != "" {
		list, err = s.attachmentSvc.GetAttachmentsBySecretVersionID(ctx, userID, svid)
	} else {
		list, err = s.attachmentSvc.GetAttachmentsBySecretID(ctx, userID, sid)
	}
	if err != nil {
		return nil, grpcError(err)
	}

	// Преобразуем вложения в proto
	out := make([]*gophkeeperpb.AttachmentSummary, 0, len(list))
	for _, a := range list {
		out = append(out, protoAttachmentSummary(a))
	}

	// Возвращаем список вложений в proto
	return &gophkeeperpb.ListAttachmentsResponse{Attachments: out}, nil
}

// GetAttachment возвращает полное вложение по id.
func (s *Server) GetAttachment(ctx context.Context, req *gophkeeperpb.GetAttachmentRequest) (*gophkeeperpb.Attachment, error) {
	// Проверяем, что пользователь авторизован
	userID, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}

	// Проверяем, что attachment_id не пустой
	if req.GetAttachmentId() == "" {
		return nil, status.Error(codes.InvalidArgument, "attachment_id required")
	}

	// Получаем вложение по id
	a, err := s.attachmentSvc.GetAttachmentByID(ctx, userID, req.GetAttachmentId())
	if err != nil {
		return nil, grpcErrorOrNotFound(err, "attachment not found")
	}

	// Преобразуем вложение в proto
	return protoAttachment(a), nil
}

// CreateAttachment создаёт вложение для указанной версии секрета.
func (s *Server) CreateAttachment(ctx context.Context, req *gophkeeperpb.CreateAttachmentRequest) (*gophkeeperpb.Attachment, error) {
	// Проверяем, что пользователь авторизован
	userID, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}

	// Проверяем, что secret_version_id не пустой
	if req.GetSecretVersionId() == "" {
		return nil, status.Error(codes.InvalidArgument, "secret_version_id required")
	}

	// Создаём вложение
	att := &model.Attachment{
		DataFormatVersion: int(req.GetDataFormatVersion()),
		InfoFormatVersion: int(req.GetInfoFormatVersion()),
		InfoEncrypted:     req.GetInfoEncrypted(),
		DataEncrypted:     req.GetDataEncrypted(),
	}
	if att.DataFormatVersion == 0 {
		att.DataFormatVersion = 1
	}
	if att.InfoFormatVersion == 0 {
		att.InfoFormatVersion = 1
	}

	// Создаём вложение
	a, err := s.attachmentSvc.CreateAttachment(ctx, userID, req.GetSecretVersionId(), att)
	if err != nil {
		return nil, grpcError(err)
	}

	// Отправляем событие аудита
	s.emitAudit(audit.ActionCreateAttachment, userID, "", a.SecretVersionID, a.ID)

	// Преобразуем вложение в proto
	return protoAttachment(a), nil
}

// DeleteAttachment удаляет вложение.
func (s *Server) DeleteAttachment(ctx context.Context, req *gophkeeperpb.DeleteAttachmentRequest) (*gophkeeperpb.DeleteAttachmentResponse, error) {
	// Проверяем, что пользователь авторизован
	userID, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}

	// Проверяем, что attachment_id не пустой
	if req.GetAttachmentId() == "" {
		return nil, status.Error(codes.InvalidArgument, "attachment_id required")
	}

	// Мягко удаляем вложение
	if err := s.attachmentSvc.HardDeleteAttachment(ctx, userID, req.GetAttachmentId()); err != nil {
		return nil, grpcError(err)
	}

	// Отправляем событие аудита
	s.emitAudit(audit.ActionDeleteAttachment, userID, "", "", req.GetAttachmentId())

	// Возвращаем пустой ответ
	return &gophkeeperpb.DeleteAttachmentResponse{}, nil
}
