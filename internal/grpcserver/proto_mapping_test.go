package grpcserver

import (
	"testing"
	"time"

	"github.com/coolycow/gophkeeper/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Конвертеры protobuf не должны паниковать на nil и сохранять поля.
func TestProtoMapping_NilPointers(t *testing.T) {
	assert.Nil(t, protoSecret(nil, nil, nil, false))
	assert.Nil(t, protoSecretVersion(nil))
	assert.Nil(t, protoAttachmentSummary(nil))
	assert.Nil(t, protoAttachment(nil))
}

func TestProtoMapping_TimeRoundTrip(t *testing.T) {
	tm := time.Unix(1700000000, 0).UTC()
	p := timeProtoPtr(&tm)
	require.NotNil(t, p)
	assert.Equal(t, tm.Unix(), p.GetSeconds())
	assert.Nil(t, timeProtoPtr(nil))
}

func TestProtoMapping_SecretWithHistory(t *testing.T) {
	s := &model.Secret{ID: "s1", UserID: "u1", CurrentSecretVersionID: "v1"}
	cur := &model.SecretVersion{ID: "v1", Version: 1}
	hist := []*model.SecretVersion{{ID: "v0", Version: 0}}
	out := protoSecret(s, cur, hist, true)
	require.NotNil(t, out)
	assert.Len(t, out.SecretVersions, 1)
}
