package grpcserver

import (
	"context"

	"github.com/coolycow/gophkeeper/internal/model"
	"github.com/coolycow/gophkeeper/internal/observer/audit"
	"github.com/coolycow/gophkeeper/internal/proto/gophkeeperpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListSecrets возвращает краткий список секретов по list_scope: активные (по умолчанию), только корзина (мягко удалённые) или все.
func (s *Server) ListSecrets(ctx context.Context, req *gophkeeperpb.ListSecretsRequest) (*gophkeeperpb.ListSecretsResponse, error) {
	// Проверяем, что пользователь авторизован
	userID, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}

	// Проверяем, что list_scope не пустой
	scope := listScopeFromProto(req.GetListScope())
	secrets, err := s.secretSvc.GetSecretsByUserID(ctx, userID, scope)
	if err != nil {
		return nil, grpcError(err)
	}

	// Преобразуем секреты в proto
	out := make([]*gophkeeperpb.SecretSummary, 0, len(secrets))
	for _, sec := range secrets {
		if sec == nil {
			continue
		}
		sum := &gophkeeperpb.SecretSummary{
			Id:                     sec.ID,
			CurrentSecretVersionId: sec.CurrentSecretVersionID,
			UpdatedAt:              timestamppb.New(*sec.UpdatedAt),
		}
		if sec.DeletedAt != nil && !sec.DeletedAt.IsZero() {
			sum.DeletedAt = timeProtoPtr(sec.DeletedAt)
		}
		out = append(out, sum)
	}

	// Возвращаем список секретов в proto
	return &gophkeeperpb.ListSecretsResponse{Secrets: out}, nil
}

// listScopeFromProto преобразует SecretListScope из proto в модель SecretListScope
func listScopeFromProto(s gophkeeperpb.SecretListScope) model.SecretListScope {
	switch s {
	case gophkeeperpb.SecretListScope_SECRET_LIST_SCOPE_ACTIVE_ONLY:
		return model.SecretListScopeActiveOnly
	case gophkeeperpb.SecretListScope_SECRET_LIST_SCOPE_DELETED_ONLY:
		return model.SecretListScopeDeletedOnly
	case gophkeeperpb.SecretListScope_SECRET_LIST_SCOPE_ALL:
		return model.SecretListScopeAll
	default:
		return model.SecretListScopeUnspecified
	}
}

// GetSecret возвращает секрет с текущей версией и опционально полной историей версий.
func (s *Server) GetSecret(ctx context.Context, req *gophkeeperpb.GetSecretRequest) (*gophkeeperpb.Secret, error) {
	// Проверяем, что пользователь авторизован
	userID, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}

	// Проверяем, что secret_id не пустой
	if req.GetSecretId() == "" {
		return nil, status.Error(codes.InvalidArgument, "secret_id required")
	}

	// Получаем секрет по ID пользователя и ID секрета
	secret, err := s.secretSvc.GetSecretByUserIDAndSecretID(ctx, userID, req.GetSecretId())
	if err != nil {
		return nil, grpcErrorOrNotFound(err, "secret not found")
	}
	if secret == nil {
		return nil, status.Error(codes.NotFound, "secret not found")
	}

	// Получаем текущую версию секрета
	cur, err := s.secretVersionSvc.GetSecretVersionByID(ctx, userID, req.GetSecretId(), secret.CurrentSecretVersionID)
	if err != nil {
		return nil, grpcErrorOrNotFound(err, "current secret version not found")
	}

	// Получаем историю версий секрета
	var hist []*model.SecretVersion
	if req.GetIncludeVersionHistory() {
		hist, err = s.secretVersionSvc.GetAllSecretHistories(ctx, userID, req.GetSecretId())
		if err != nil {
			return nil, grpcError(err)
		}
	}

	return protoSecret(secret, cur, hist, req.GetIncludeVersionHistory()), nil
}

// CreateSecret создаёт секрет с первой версией ciphertext.
func (s *Server) CreateSecret(ctx context.Context, req *gophkeeperpb.CreateSecretRequest) (*gophkeeperpb.Secret, error) {
	// Проверяем, что пользователь авторизован
	userID, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}

	// Создаём секрет
	sec, err := s.secretSvc.CreateSecret(ctx, userID, model.SecretCreateRequest{
		DataEncrypted:     req.GetDataEncrypted(),
		DataFormatVersion: int(req.GetDataFormatVersion()),
	})

	// Если ошибка при создании секрета, возвращаем ошибку
	if err != nil {
		return nil, grpcError(err)
	}

	// Если количество версий секрета равно 0, возвращаем ошибку
	if len(sec.SecretVersions) == 0 {
		return nil, status.Error(codes.Internal, "secret created without version")
	}

	// Отправляем событие аудита
	s.emitAudit(audit.ActionCreateSecret, userID, sec.ID, sec.CurrentSecretVersionID, "")

	// Преобразуем секрет в proto
	return protoSecret(sec, sec.SecretVersions[0], nil, false), nil
}

// UpdateSecret добавляет новую версию секрета и делает её текущей.
func (s *Server) UpdateSecret(ctx context.Context, req *gophkeeperpb.UpdateSecretRequest) (*gophkeeperpb.SecretVersion, error) {
	// Проверяем, что пользователь авторизован
	userID, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}

	// Проверяем, что secret_id не пустой
	if req.GetSecretId() == "" {
		return nil, status.Error(codes.InvalidArgument, "secret_id required")
	}

	// Проверяем, что data_format_version не пустой
	dfv := int(req.GetDataFormatVersion())
	if dfv == 0 {
		dfv = 1
	}

	// Создаём версию секрета
	sv, err := s.secretVersionSvc.CreateSecretVersion(ctx, userID, req.GetSecretId(), &model.SecretVersion{
		DataEncrypted:     req.GetDataEncrypted(),
		DataFormatVersion: dfv,
	})

	// Если ошибка при создании версии секрета, возвращаем ошибку
	if err != nil {
		return nil, grpcError(err)
	}

	// Отправляем событие аудита
	s.emitAudit(audit.ActionCreateSecretVersion, userID, req.GetSecretId(), sv.ID, "")

	// Преобразуем версию секрета в proto

	// Преобразуем версию секрета в proto
	return protoSecretVersion(sv), nil
}

// DeleteSecret мягко удаляет секрет.
func (s *Server) DeleteSecret(ctx context.Context, req *gophkeeperpb.DeleteSecretRequest) (*gophkeeperpb.DeleteSecretResponse, error) {
	// Проверяем, что пользователь авторизован
	userID, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}

	// Проверяем, что secret_id не пустой
	if req.GetSecretId() == "" {
		return nil, status.Error(codes.InvalidArgument, "secret_id required")
	}

	// Мягко удаляем секрет
	if err := s.secretSvc.SoftDeleteSecret(ctx, userID, req.GetSecretId()); err != nil {
		return nil, grpcError(err)
	}

	// Отправляем событие аудита
	s.emitAudit(audit.ActionDeleteSecret, userID, req.GetSecretId(), "", "")

	// Возвращаем пустой ответ
	return &gophkeeperpb.DeleteSecretResponse{}, nil
}
