package clientgrpc

import (
	"testing"
	"time"

	"github.com/coolycow/gophkeeper/internal/proto/gophkeeperpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestClient_ApplyAuthResponse(t *testing.T) {
	c := NewClient(nil)
	exp := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	err := c.ApplyAuthResponse(&gophkeeperpb.AuthResponse{
		Token:        "access",
		RefreshToken: "refresh",
		Salt:         "deadbeef",
		ExpiresAt:    timestamppb.New(exp),
	})
	require.NoError(t, err)
	assert.Equal(t, "deadbeef", c.SaltHex())
	assert.Equal(t, "refresh", c.RefreshToken())
	assert.Equal(t, exp.UTC(), c.AccessExpiresAt().UTC())
}

func TestClient_ApplyAuthResponse_nil(t *testing.T) {
	c := NewClient(nil)
	require.Error(t, c.ApplyAuthResponse(nil))
}

func TestClient_ApplyAuthResponse_missingTokens(t *testing.T) {
	c := NewClient(nil)
	err := c.ApplyAuthResponse(&gophkeeperpb.AuthResponse{Salt: "x"})
	require.Error(t, err)
}
