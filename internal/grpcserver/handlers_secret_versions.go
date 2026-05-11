package grpcserver

import (
	"context"

	"github.com/coolycow/gophkeeper/internal/observer/audit"
	"github.com/coolycow/gophkeeper/internal/proto/gophkeeperpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListSecretVersions возвращает все версии секрета.
func (s *Server) ListSecretVersions(ctx context.Context, req *gophkeeperpb.ListSecretVersionsRequest) (*gophkeeperpb.ListSecretVersionsResponse, error) {
	// Проверяем, что пользователь авторизован
	userID, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}

	// Проверяем, что secret_id не пустой
	if req.GetSecretId() == "" {
		return nil, status.Error(codes.InvalidArgument, "secret_id required")
	}

	// Получаем все версии секрета
	vers, err := s.secretVersionSvc.GetAllSecretHistories(ctx, userID, req.GetSecretId())
	if err != nil {
		return nil, grpcError(err)
	}

	// Преобразуем версии секрета в proto
	out := make([]*gophkeeperpb.SecretVersion, 0, len(vers))
	for _, v := range vers {
		out = append(out, protoSecretVersion(v))
	}

	// Возвращаем все версии секрета в proto
	return &gophkeeperpb.ListSecretVersionsResponse{Versions: out}, nil
}

// GetSecretVersion возвращает одну версию по идентификатору.
func (s *Server) GetSecretVersion(ctx context.Context, req *gophkeeperpb.GetSecretVersionRequest) (*gophkeeperpb.SecretVersion, error) {
	// Проверяем, что пользователь авторизован
	userID, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}

	// Проверяем, что secret_id и secret_version_id не пустые
	if req.GetSecretId() == "" || req.GetSecretVersionId() == "" {
		return nil, status.Error(codes.InvalidArgument, "secret_id and secret_version_id required")
	}

	// Получаем версию секрета по идентификатору
	v, err := s.secretVersionSvc.GetSecretVersionByID(ctx, userID, req.GetSecretId(), req.GetSecretVersionId())

	// Если версия секрета не найдена, возвращаем ошибку
	if err != nil {
		return nil, grpcErrorOrNotFound(err, "secret version not found")
	}

	// Преобразуем версию секрета в proto
	return protoSecretVersion(v), nil
}

// DeleteSecretVersion удаляет версию секрета (только жёсткое удаление).
func (s *Server) DeleteSecretVersion(ctx context.Context, req *gophkeeperpb.DeleteSecretVersionRequest) (*gophkeeperpb.DeleteSecretVersionResponse, error) {
	// Проверяем, что пользователь авторизован
	userID, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}

	// Проверяем, что secret_id и secret_version_id не пустые
	if req.GetSecretId() == "" || req.GetSecretVersionId() == "" {
		return nil, status.Error(codes.InvalidArgument, "secret_id and secret_version_id required")
	}

	// Удаляем версию секрета (только жёсткое удаление)
	if err := s.secretVersionSvc.HardDeleteSecretVersion(ctx, userID, req.GetSecretId(), req.GetSecretVersionId()); err != nil {
		return nil, grpcError(err)
	}

	// Отправляем событие аудита
	s.emitAudit(audit.ActionDeleteSecretVersion, userID, req.GetSecretId(), req.GetSecretVersionId(), "")

	// Возвращаем идентификатор удаленной версии секрета
	return &gophkeeperpb.DeleteSecretVersionResponse{SecretVersionId: req.GetSecretVersionId()}, nil
}

// RestoreSecretVersion восстанавливает версию секрета (устанавливает её текущей для секрета).
func (s *Server) RestoreSecretVersion(ctx context.Context, req *gophkeeperpb.RestoreSecretVersionRequest) (*gophkeeperpb.RestoreSecretVersionResponse, error) {
	// Проверяем, что пользователь авторизован
	userID, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}

	// Проверяем, что secret_id и secret_version_id не пустые
	if req.GetSecretId() == "" || req.GetSecretVersionId() == "" {
		return nil, status.Error(codes.InvalidArgument, "secret_id and secret_version_id required")
	}

	// Восстанавливаем версию секрета (устанавливает её текущей для секрета)
	if err := s.secretVersionSvc.RestoreSecretVersion(ctx, userID, req.GetSecretId(), req.GetSecretVersionId()); err != nil {
		return nil, grpcError(err)
	}

	// Отправляем событие аудита
	s.emitAudit(audit.ActionUpdateSecretVersion, userID, req.GetSecretId(), req.GetSecretVersionId(), "")

	// Возвращаем идентификатор восстановленной версии секрета
	return &gophkeeperpb.RestoreSecretVersionResponse{SecretVersionId: req.GetSecretVersionId()}, nil
}
