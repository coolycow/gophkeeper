package grpcserver

import (
	"database/sql"
	"errors"
	"net/http"
	"testing"

	httperr "github.com/coolycow/gophkeeper/internal/error"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Базовый маппинг CustomError в gRPC-коды.
func TestGRPCError_CustomErrorVariants(t *testing.T) {
	cases := []struct {
		code int
		want codes.Code
	}{
		{http.StatusBadRequest, codes.InvalidArgument},
		{http.StatusNotFound, codes.NotFound},
		{http.StatusConflict, codes.AlreadyExists},
		{http.StatusUnauthorized, codes.Unauthenticated},
		{http.StatusTeapot, codes.Internal},
	}

	for _, tc := range cases {
		se, ok := status.FromError(grpcError(httperr.CustomError{
			Message:    "msg",
			StatusCode: tc.code,
		}))
		require.True(t, ok)
		assert.Equal(t, tc.want, se.Code())
		assert.Equal(t, "msg", se.Message())
	}
}

func TestGRPCError_Nil(t *testing.T) {
	assert.Nil(t, grpcError(nil))
}

func TestGRPCError_NonCustom(t *testing.T) {
	e := grpcError(errors.New("plain"))
	se, ok := status.FromError(e)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, se.Code())
}

// sql.ErrNoRows → NotFound через grpcErrorOrNotFound.
func TestGRPCErrorOrNotFound_ErrNoRows(t *testing.T) {
	e := grpcErrorOrNotFound(sql.ErrNoRows, "gone")
	se, ok := status.FromError(e)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, se.Code())
}
