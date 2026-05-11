package clientsession

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSaveLoad_roundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.json")

	f := &File{ServerAddress: "127.0.0.1:8081", Email: "a@b.c", RefreshToken: "rt-secret"}
	require.NoError(t, Save(path, f))

	got, err := Load(path)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, f.ServerAddress, got.ServerAddress)
	require.Equal(t, f.Email, got.Email)
	require.Equal(t, f.RefreshToken, got.RefreshToken)
}

func TestLoad_missing(t *testing.T) {
	got, err := Load(filepath.Join(t.TempDir(), "nope.json"))
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestDefaultPath_containsAppName(t *testing.T) {
	p, err := DefaultPath()
	require.NoError(t, err)
	require.Contains(t, p, "gophkeeper")
}

func TestClear(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.json")
	require.NoError(t, Save(path, &File{RefreshToken: "x"}))
	require.NoError(t, Clear(path))
	got, err := Load(path)
	require.NoError(t, err)
	require.Nil(t, got)
}
