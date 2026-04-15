package shortener_test

import (
	"testing"

	"github.com/coolycow/shortener/internal/shortener"
	"github.com/stretchr/testify/require"
)

// TestNormalizeShortenInput_Invalid проверяет отклонение пустой строки и невалидного URL.
func TestNormalizeShortenInput_Invalid(t *testing.T) {
	t.Parallel()

	_, err := shortener.NormalizeShortenInput("")
	require.Error(t, err)

	_, err = shortener.NormalizeShortenInput("not-a-url")
	require.Error(t, err)
}
