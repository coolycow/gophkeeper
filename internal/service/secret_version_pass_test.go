package service

import (
	"context"
	"testing"

	"github.com/coolycow/gophkeeper/internal/config"
	"github.com/coolycow/gophkeeper/internal/model"
	"github.com/coolycow/gophkeeper/internal/repository/repotest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Делегирование CRUD версий в репозиторий.
func TestSecretVersionService_Passthrough(t *testing.T) {
	repo := &repotest.Stub{
		GetLatestSecretVersionFn: func(_ context.Context, u, sid string) (*model.SecretVersion, error) {
			return &model.SecretVersion{ID: "v-last", SecretID: sid}, nil
		},
		GetSecretVersionByIDFn: func(_ context.Context, u, sid, vid string) (*model.SecretVersion, error) {
			return &model.SecretVersion{ID: vid}, nil
		},
		GetAllSecretHistoriesFn: func(context.Context, string, string) ([]*model.SecretVersion, error) {
			return []*model.SecretVersion{{Version: 1}}, nil
		},
		HardDeleteSecretVersionFn: func(context.Context, string, string, string) error {
			return nil
		},
		RestoreSecretVersionFn: func(context.Context, string, string, string) error {
			return nil
		},
	}
	svc := NewSecretVersionService(&config.ConfigServer{}, repo)
	ctx := context.Background()

	v, err := svc.GetLatestSecretVersion(ctx, "u", "s")
	require.NoError(t, err)
	assert.Equal(t, "v-last", v.ID)

	v2, err := svc.GetSecretVersionByID(ctx, "u", "s", "v99")
	require.NoError(t, err)
	assert.Equal(t, "v99", v2.ID)

	h, err := svc.GetAllSecretHistories(ctx, "u", "s")
	require.NoError(t, err)
	require.Len(t, h, 1)

	require.NoError(t, svc.HardDeleteSecretVersion(ctx, "u", "s", "v99"))
	require.NoError(t, svc.RestoreSecretVersion(ctx, "u", "s", "v99"))
}
