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

// Вложения: проксирование и подстановка версий формата по умолчанию.
func TestAttachmentService_Passthrough(t *testing.T) {
	repo := &repotest.Stub{
		GetAttachmentByIDFn: func(_ context.Context, u, id string) (*model.Attachment, error) {
			return &model.Attachment{ID: id, SecretVersionID: u}, nil
		},
		GetAttachmentsBySecretIDFn: func(context.Context, string, string) ([]*model.AttachmentSummary, error) {
			return []*model.AttachmentSummary{{ID: "x"}}, nil
		},
		GetAttachmentsBySecretVersionIDFn: func(context.Context, string, string) ([]*model.AttachmentSummary, error) {
			return nil, nil
		},
		HardDeleteAttachmentFn: func(context.Context, string, string) error { return nil },
	}
	svc := NewAttachmentService(&config.ConfigServer{}, repo)
	ctx := context.Background()

	a, err := svc.GetAttachmentByID(ctx, "su", "aid")
	require.NoError(t, err)
	assert.Equal(t, "aid", a.ID)

	list, err := svc.GetAttachmentsBySecretID(ctx, "u", "s")
	require.NoError(t, err)
	require.Len(t, list, 1)

	require.NoError(t, svc.HardDeleteAttachment(ctx, "u", "a"))
}

func TestAttachmentService_CreateAttachment_DefaultVersions(t *testing.T) {
	repo := &repotest.Stub{
		CreateAttachmentFn: func(_ context.Context, _ string, sv string, att *model.Attachment) (*model.Attachment, error) {
			assert.Equal(t, 1, att.DataFormatVersion)
			assert.Equal(t, 1, att.InfoFormatVersion)
			return att, nil
		},
	}
	svc := NewAttachmentService(&config.ConfigServer{}, repo)
	_, err := svc.CreateAttachment(context.Background(), "u", "sv", &model.Attachment{
		InfoEncrypted: []byte("i"),
		DataEncrypted: []byte("d"),
	})
	require.NoError(t, err)
}
