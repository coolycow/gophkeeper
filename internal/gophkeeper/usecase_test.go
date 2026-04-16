package gophkeeper_test

import (
	"testing"

	"github.com/coolycow/gophkeeper/internal/gophkeeper"
	"github.com/stretchr/testify/require"
)

// TestNormalizeShortenInput_Invalid проверяет отклонение пустой строки и невалидного URL.
func TestNormalizeShortenInput_Invalid(t *testing.T) {
	t.Parallel()

	_, err := gophkeeper.NormalizeShortenInput("")
	require.Error(t, err)

	_, err = gophkeeper.NormalizeShortenInput("not-a-url")
	require.Error(t, err)
}
