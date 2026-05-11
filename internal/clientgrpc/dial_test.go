package clientgrpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDial_emptyAddress(t *testing.T) {
	_, err := Dial(context.Background(), DialOptions{})
	require.Error(t, err)
}
