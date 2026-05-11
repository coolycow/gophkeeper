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

// CreateSecret подставляет data_format_version = 1 и передаёт соль версии из запроса.
func TestSecretService_CreateSecret_DefaultVersion(t *testing.T) {
	var seen *model.Secret
	repo := &repotest.Stub{
		CreateSecretFn: func(_ context.Context, userID string, sec *model.Secret) (*model.Secret, error) {
			assert.Equal(t, "usr", userID)
			require.Len(t, sec.SecretVersions, 1)
			assert.Equal(t, 1, sec.SecretVersions[0].DataFormatVersion)
			seen = sec
			return &model.Secret{ID: "sid"}, nil
		},
	}

	svc := NewSecretService(&config.ConfigServer{}, repo)

	out, err := svc.CreateSecret(context.Background(), "usr", model.SecretCreateRequest{
		DataEncrypted: []byte("x"),
	})
	require.NoError(t, err)
	assert.Equal(t, "sid", out.ID)
	require.NotNil(t, seen)
	assert.Nil(t, seen.DeletedAt)
}
