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

// Проверки «тонкого» слоя: делегирование в repo без лишней логики.
func TestUserService_GetUserByID_Delegates(t *testing.T) {
	repo := &repotest.Stub{
		GetUserByIDFn: func(_ context.Context, id string) (*model.User, error) {
			return &model.User{ID: id}, nil
		},
	}
	u, err := NewUserService(&config.ConfigServer{}, repo).GetUserByID(context.Background(), "u1")
	require.NoError(t, err)
	assert.Equal(t, "u1", u.ID)
}

func TestUserService_GetUserByEmail_Valid(t *testing.T) {
	repo := &repotest.Stub{
		GetUserByEmailFn: func(_ context.Context, email string) (*model.User, error) {
			assert.Equal(t, "a@example.com", email)
			return &model.User{Email: email}, nil
		},
	}
	u, err := NewUserService(&config.ConfigServer{}, repo).GetUserByEmail(context.Background(), "a@example.com")
	require.NoError(t, err)
	assert.Equal(t, "a@example.com", u.Email)
}

func TestUserService_DeleteUsers_Delegate(t *testing.T) {
	var soft, hard string
	repo := &repotest.Stub{
		SoftDeleteUserFn: func(_ context.Context, id string) error {
			soft = id
			return nil
		},
		HardDeleteUserFn: func(_ context.Context, id string) error {
			hard = id
			return nil
		},
	}
	svc := NewUserService(&config.ConfigServer{}, repo)
	require.NoError(t, svc.SoftDeleteUser(context.Background(), "s1"))
	require.NoError(t, svc.HardDeleteUser(context.Background(), "h1"))
	assert.Equal(t, "s1", soft)
	assert.Equal(t, "h1", hard)
}

func TestSecretService_ReadMethods_Delegate(t *testing.T) {
	repo := &repotest.Stub{
		GetSecretByIDFn: func(_ context.Context, id string) (*model.Secret, error) {
			return &model.Secret{ID: id}, nil
		},
		GetSecretsByUserIDFn: func(_ context.Context, uid string, scope model.SecretListScope) ([]*model.Secret, error) {
			return []*model.Secret{{UserID: uid}}, nil
		},
		GetSecretByUserIDAndSecretIDFn: func(_ context.Context, uid, sid string) (*model.Secret, error) {
			return &model.Secret{UserID: uid, ID: sid}, nil
		},
	}
	cfg := &config.ConfigServer{}
	svc := NewSecretService(cfg, repo)

	s1, err := svc.GetSecretByID(context.Background(), "x")
	require.NoError(t, err)
	assert.Equal(t, "x", s1.ID)

	ss, err := svc.GetSecretsByUserID(context.Background(), "u", model.SecretListScopeActiveOnly)
	require.NoError(t, err)
	require.Len(t, ss, 1)

	s2, err := svc.GetSecretByUserIDAndSecretID(context.Background(), "u", "sid")
	require.NoError(t, err)
	assert.Equal(t, "sid", s2.ID)
}
