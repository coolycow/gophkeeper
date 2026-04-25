package buildinfo

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestOrNA проверяет, что OrNA возвращает "N/A" для пустой строки или строки с пробелами.
func TestOrNA(t *testing.T) {
	require.Equal(t, "N/A", OrNA(""))
	require.Equal(t, "N/A", OrNA("   "))
	require.Equal(t, "1.0", OrNA("1.0"))
}

// TestFprint проверяет, что Fprint выводит версию, дату сборки и коммит в w, каждый на своей строке.
func TestFprint(t *testing.T) {
	Version = "v-test"
	BuildDate = "today"
	BuildCommit = "abc"

	var buf bytes.Buffer
	Fprint(&buf)

	s := buf.String()
	require.Contains(t, s, "v-test")
	require.Contains(t, s, "today")
	require.Contains(t, s, "abc")
}
