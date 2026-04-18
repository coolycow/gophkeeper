package grpcserver

import (
	"context"

	"github.com/coolycow/gophkeeper/internal/model"
	"github.com/coolycow/gophkeeper/internal/proto/gophkeeperpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListSecrets возвращает краткий список секретов по list_scope: активные (по умолчанию), только корзина (мягко удалённые) или все.
func (s *Server) ListSecrets(ctx context.Context, req *gophkeeperpb.ListSecretsRequest) (*gophkeeperpb.ListSecretsResponse, error) {
	userID, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}

	scope := listScopeFromProto(req.GetListScope())
	secrets, err := s.secretSvc.GetSecretsByUserID(ctx, userID, scope)
	if err != nil {
		return nil, grpcError(err)
	}

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
		if !sec.DeletedAt.IsZero() {
			sum.DeletedAt = timestamppb.New(*sec.DeletedAt)
		}
		out = append(out, sum)
	}

	return &gophkeeperpb.ListSecretsResponse{Secrets: out}, nil
}

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
	userID, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetSecretId() == "" {
		return nil, status.Error(codes.InvalidArgument, "secret_id required")
	}

	secret, err := s.secretSvc.GetSecretByUserIDAndSecretID(ctx, userID, req.GetSecretId())
	if err != nil {
		return nil, grpcErrorOrNotFound(err, "secret not found")
	}
	if secret == nil {
		return nil, status.Error(codes.NotFound, "secret not found")
	}

	cur, err := s.secretVersionSvc.GetSecretVersionByID(ctx, userID, req.GetSecretId(), secret.CurrentSecretVersionID)
	if err != nil {
		return nil, grpcErrorOrNotFound(err, "current secret version not found")
	}

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
	userID, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}

	sec, err := s.secretSvc.CreateSecret(ctx, userID, model.SecretCreateRequest{
		DataEncrypted:     req.GetDataEncrypted(),
		DataFormatVersion: int(req.GetDataFormatVersion()),
	})
	if err != nil {
		return nil, grpcError(err)
	}
	if len(sec.SecretVersions) == 0 {
		return nil, status.Error(codes.Internal, "secret created without version")
	}
	return protoSecret(sec, sec.SecretVersions[0], nil, false), nil
}

// UpdateSecret добавляет новую версию секрета и делает её текущей.
func (s *Server) UpdateSecret(ctx context.Context, req *gophkeeperpb.UpdateSecretRequest) (*gophkeeperpb.SecretVersion, error) {
	userID, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetSecretId() == "" {
		return nil, status.Error(codes.InvalidArgument, "secret_id required")
	}

	dfv := int(req.GetDataFormatVersion())
	if dfv == 0 {
		dfv = 1
	}

	sv, err := s.secretVersionSvc.CreateSecretVersion(ctx, userID, req.GetSecretId(), &model.SecretVersion{
		DataEncrypted:     req.GetDataEncrypted(),
		DataFormatVersion: dfv,
	})
	if err != nil {
		return nil, grpcError(err)
	}
	return protoSecretVersion(sv), nil
}

// DeleteSecret мягко удаляет секрет.
func (s *Server) DeleteSecret(ctx context.Context, req *gophkeeperpb.DeleteSecretRequest) (*gophkeeperpb.DeleteSecretResponse, error) {
	userID, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetSecretId() == "" {
		return nil, status.Error(codes.InvalidArgument, "secret_id required")
	}

	if err := s.secretSvc.SoftDeleteSecret(ctx, userID, req.GetSecretId()); err != nil {
		return nil, grpcError(err)
	}
	return &gophkeeperpb.DeleteSecretResponse{}, nil
}

// ListSecretVersions возвращает все версии секрета.
func (s *Server) ListSecretVersions(ctx context.Context, req *gophkeeperpb.ListSecretVersionsRequest) (*gophkeeperpb.ListSecretVersionsResponse, error) {
	userID, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetSecretId() == "" {
		return nil, status.Error(codes.InvalidArgument, "secret_id required")
	}

	vers, err := s.secretVersionSvc.GetAllSecretHistories(ctx, userID, req.GetSecretId())
	if err != nil {
		return nil, grpcError(err)
	}
	out := make([]*gophkeeperpb.SecretVersion, 0, len(vers))
	for _, v := range vers {
		out = append(out, protoSecretVersion(v))
	}
	return &gophkeeperpb.ListSecretVersionsResponse{Versions: out}, nil
}

// GetSecretVersion возвращает одну версию по идентификатору.
func (s *Server) GetSecretVersion(ctx context.Context, req *gophkeeperpb.GetSecretVersionRequest) (*gophkeeperpb.SecretVersion, error) {
	userID, err := requireUserID(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetSecretId() == "" || req.GetSecretVersionId() == "" {
		return nil, status.Error(codes.InvalidArgument, "secret_id and secret_version_id required")
	}

	v, err := s.secretVersionSvc.GetSecretVersionByID(ctx, userID, req.GetSecretId(), req.GetSecretVersionId())
	if err != nil {
		return nil, grpcErrorOrNotFound(err, "secret version not found")
	}
	return protoSecretVersion(v), nil
}
