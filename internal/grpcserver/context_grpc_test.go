package grpcserver

import (
	"context"
	"testing"

	"github.com/coolycow/gophkeeper/internal/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// requireUserID без user в контексте → Unauthenticated.
func TestRequireUserID_Missing(t *testing.T) {
	ctx := context.Background()
	_, err := requireUserID(ctx)
	se, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, se.Code())

	ctx = ctxutil.WithUserID(ctx, "u1")
	id, err := requireUserID(ctx)
	require.NoError(t, err)
	assert.Equal(t, "u1", id)
}
