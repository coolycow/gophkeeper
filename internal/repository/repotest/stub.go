// Package repotest содержит заглушки для тестов над GophKeeperRepository.
package repotest

import (
	"context"
	"time"

	"github.com/coolycow/gophkeeper/internal/model"
	"github.com/coolycow/gophkeeper/internal/repository"
)

var _ repository.GophKeeperRepository = (*Stub)(nil)

// Stub заглушка: по умолчанию безопасные нули; переопределяйте только нужные *Fn-поля.
type Stub struct {
	CloseErr error

	CreateSecretFn func(ctx context.Context, userID string, secret *model.Secret) (*model.Secret, error)
	UpdateSecretFn func(ctx context.Context, userID, secretID string, secret *model.Secret) error
	GetSecretByUserIDAndSecretIDFn func(ctx context.Context, userID, secretID string) (*model.Secret, error)

	PingFn             func(context.Context) error
	RunMigrationsFn    func() error
	GetUserByIDFn      func(context.Context, string) (*model.User, error)
	GetUserByEmailFn   func(context.Context, string) (*model.User, error)
	SoftDeleteUserFn   func(context.Context, string) error
	HardDeleteUserFn   func(context.Context, string) error
	GetSecretByIDFn    func(context.Context, string) (*model.Secret, error)
	GetSecretsByUserIDFn func(context.Context, string, model.SecretListScope) ([]*model.Secret, error)

	GetLatestSecretVersionFn   func(context.Context, string, string) (*model.SecretVersion, error)
	GetSecretVersionByIDFn      func(context.Context, string, string, string) (*model.SecretVersion, error)
	GetAllSecretHistoriesFn      func(context.Context, string, string) ([]*model.SecretVersion, error)
	CreateSecretVersionFn        func(context.Context, string, string, *model.SecretVersion) (*model.SecretVersion, error)
	HardDeleteOldestSecretVersionFn func(context.Context, string, string) error
	HardDeleteSecretVersionFn    func(context.Context, string, string, string) error
	RestoreSecretVersionFn       func(context.Context, string, string, string) error

	GetAttachmentByIDFn               func(context.Context, string, string) (*model.Attachment, error)
	GetAttachmentsBySecretIDFn        func(context.Context, string, string) ([]*model.AttachmentSummary, error)
	GetAttachmentsBySecretVersionIDFn func(context.Context, string, string) ([]*model.AttachmentSummary, error)
	CreateAttachmentFn                func(context.Context, string, string, *model.Attachment) (*model.Attachment, error)
	HardDeleteAttachmentFn            func(context.Context, string, string) error
}

func (s *Stub) Close() error                                                       { return s.CloseErr }
func (s *Stub) Ping(ctx context.Context) error                                       { return call(s.PingFn, ctx) }
func (s *Stub) RunMigrations() error                                                 { return call0(s.RunMigrationsFn) }
func (s *Stub) GetUserByID(ctx context.Context, id string) (*model.User, error)    { return callUser(s.GetUserByIDFn, ctx, id) }
func (s *Stub) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	if s.GetUserByEmailFn != nil {
		return s.GetUserByEmailFn(ctx, email)
	}
	return nil, nil
}
func (s *Stub) CreateUser(context.Context, *model.User) (*model.User, error)       { return nil, nil }
func (s *Stub) UpdateUser(context.Context, string, *model.User) error               { return nil }
func (s *Stub) SoftDeleteUser(ctx context.Context, id string) error {
	if s.SoftDeleteUserFn != nil {
		return s.SoftDeleteUserFn(ctx, id)
	}
	return nil
}
func (s *Stub) HardDeleteUser(ctx context.Context, id string) error {
	if s.HardDeleteUserFn != nil {
		return s.HardDeleteUserFn(ctx, id)
	}
	return nil
}
func (s *Stub) GetUsersCount(context.Context) int                                   { return 0 }
func (s *Stub) CreateRefreshToken(context.Context, string, []byte, time.Time) (*model.RefreshToken, error) {
	return nil, nil
}
func (s *Stub) DeleteRefreshToken(context.Context, string) error { return nil }
func (s *Stub) FindValidRefreshTokenByHash(context.Context, []byte, time.Time) (*model.RefreshToken, error) {
	return nil, nil
}

func (s *Stub) GetSecretByID(ctx context.Context, secretID string) (*model.Secret, error) {
	if s.GetSecretByIDFn != nil {
		return s.GetSecretByIDFn(ctx, secretID)
	}
	return nil, nil
}

func (s *Stub) GetSecretsByUserID(ctx context.Context, userID string, scope model.SecretListScope) ([]*model.Secret, error) {
	if s.GetSecretsByUserIDFn != nil {
		return s.GetSecretsByUserIDFn(ctx, userID, scope)
	}
	return nil, nil
}

func (s *Stub) GetSecretByUserIDAndSecretID(ctx context.Context, uid, sid string) (*model.Secret, error) {
	if s.GetSecretByUserIDAndSecretIDFn != nil {
		return s.GetSecretByUserIDAndSecretIDFn(ctx, uid, sid)
	}
	return nil, nil
}

func (s *Stub) CreateSecret(ctx context.Context, uid string, secret *model.Secret) (*model.Secret, error) {
	if s.CreateSecretFn != nil {
		return s.CreateSecretFn(ctx, uid, secret)
	}
	return secret, nil
}

func (s *Stub) UpdateSecret(ctx context.Context, uid, sid string, secret *model.Secret) error {
	if s.UpdateSecretFn != nil {
		return s.UpdateSecretFn(ctx, uid, sid, secret)
	}
	return nil
}

func (s *Stub) SoftDeleteSecret(context.Context, string, string) error   { return nil }
func (s *Stub) HardDeleteSecret(context.Context, string, string) error    { return nil }
func (s *Stub) CompressSecretByID(context.Context, string, string) error { return nil }
func (s *Stub) CompressSecretsByUserID(context.Context, string) error   { return nil }
func (s *Stub) GetMaxSecretVersion(context.Context, string, string) (int, error) {
	return 0, nil
}

func (s *Stub) GetCurrentSecretVersion(context.Context, string, string) (*model.SecretVersion, error) {
	return nil, nil
}
func (s *Stub) GetLatestSecretVersion(ctx context.Context, uid, sid string) (*model.SecretVersion, error) {
	if s.GetLatestSecretVersionFn != nil {
		return s.GetLatestSecretVersionFn(ctx, uid, sid)
	}
	return nil, nil
}
func (s *Stub) GetSecretVersionByID(ctx context.Context, uid, sid, vid string) (*model.SecretVersion, error) {
	if s.GetSecretVersionByIDFn != nil {
		return s.GetSecretVersionByIDFn(ctx, uid, sid, vid)
	}
	return nil, nil
}
func (s *Stub) GetAllSecretHistories(ctx context.Context, uid, sid string) ([]*model.SecretVersion, error) {
	if s.GetAllSecretHistoriesFn != nil {
		return s.GetAllSecretHistoriesFn(ctx, uid, sid)
	}
	return nil, nil
}
func (s *Stub) CreateSecretVersion(ctx context.Context, uid, sid string, v *model.SecretVersion) (*model.SecretVersion, error) {
	if s.CreateSecretVersionFn != nil {
		return s.CreateSecretVersionFn(ctx, uid, sid, v)
	}
	return v, nil
}
func (s *Stub) HardDeleteSecretVersion(ctx context.Context, uid, sid, vid string) error {
	if s.HardDeleteSecretVersionFn != nil {
		return s.HardDeleteSecretVersionFn(ctx, uid, sid, vid)
	}
	return nil
}
func (s *Stub) HardDeleteOldestSecretVersion(ctx context.Context, uid, sid string) error {
	if s.HardDeleteOldestSecretVersionFn != nil {
		return s.HardDeleteOldestSecretVersionFn(ctx, uid, sid)
	}
	return nil
}
func (s *Stub) RestoreSecretVersion(ctx context.Context, uid, sid, vid string) error {
	if s.RestoreSecretVersionFn != nil {
		return s.RestoreSecretVersionFn(ctx, uid, sid, vid)
	}
	return nil
}

func (s *Stub) GetAttachmentByID(ctx context.Context, uid, aid string) (*model.Attachment, error) {
	if s.GetAttachmentByIDFn != nil {
		return s.GetAttachmentByIDFn(ctx, uid, aid)
	}
	return nil, nil
}

func (s *Stub) GetAttachmentsBySecretID(ctx context.Context, uid, sid string) ([]*model.AttachmentSummary, error) {
	if s.GetAttachmentsBySecretIDFn != nil {
		return s.GetAttachmentsBySecretIDFn(ctx, uid, sid)
	}
	return nil, nil
}

func (s *Stub) GetAttachmentsBySecretVersionID(ctx context.Context, uid, vid string) ([]*model.AttachmentSummary, error) {
	if s.GetAttachmentsBySecretVersionIDFn != nil {
		return s.GetAttachmentsBySecretVersionIDFn(ctx, uid, vid)
	}
	return nil, nil
}

func (s *Stub) CreateAttachment(ctx context.Context, uid, vid string, a *model.Attachment) (*model.Attachment, error) {
	if s.CreateAttachmentFn != nil {
		return s.CreateAttachmentFn(ctx, uid, vid, a)
	}
	return a, nil
}

func (s *Stub) HardDeleteAttachment(ctx context.Context, uid, aid string) error {
	if s.HardDeleteAttachmentFn != nil {
		return s.HardDeleteAttachmentFn(ctx, uid, aid)
	}
	return nil
}

func call(fn func(context.Context) error, ctx context.Context) error {
	if fn != nil {
		return fn(ctx)
	}
	return nil
}

func call0(fn func() error) error {
	if fn != nil {
		return fn()
	}
	return nil
}

func callUser(fn func(context.Context, string) (*model.User, error), ctx context.Context, id string) (*model.User, error) {
	if fn != nil {
		return fn(ctx, id)
	}
	return nil, nil
}
