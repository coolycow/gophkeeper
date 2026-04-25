package secretcrypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt_roundTrip(t *testing.T) {
	saltHex := "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"
	password := "my-secret-password"
	plain := []byte(`{"kind":"text","text":"hello"}`)

	ct, err := Encrypt(plain, password, saltHex)
	require.NoError(t, err)
	require.Greater(t, len(ct), magicLen+nonceLen)
	require.Equal(t, magic, string(ct[:magicLen]))

	got, err := Decrypt(ct, password, saltHex)
	require.NoError(t, err)
	require.True(t, bytes.Equal(plain, got))
}

func TestDecrypt_wrongPassword(t *testing.T) {
	saltHex := "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"
	ct, err := Encrypt([]byte("x"), "right", saltHex)
	require.NoError(t, err)

	_, err = Decrypt(ct, "wrong", saltHex)
	require.ErrorIs(t, err, ErrDecrypt)
}

func TestDecrypt_truncated(t *testing.T) {
	_, err := Decrypt([]byte("GK"), "p", "ab")
	require.Error(t, err)
}

func TestDecrypt_badMagic(t *testing.T) {
	saltHex := "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"
	b := make([]byte, magicLen+nonceLen+8)
	copy(b, []byte("XXXX"))

	_, err := Decrypt(b, "p", saltHex)
	require.ErrorIs(t, err, ErrDecrypt)
}

func TestDataFormatVersion_constant(t *testing.T) {
	require.Equal(t, 1, DataFormatVersion)
}

func TestDeriveKey_EncryptWithKey_roundTrip(t *testing.T) {
	saltHex := "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"
	password := "batch-decrypt"
	plain := []byte("list title")
	key, err := DeriveKeyFromPassword(password, saltHex)
	require.NoError(t, err)
	require.Len(t, key, 32)
	ct, err := EncryptWithKey(plain, key)
	require.NoError(t, err)
	got, err := DecryptWithKey(ct, key)
	require.NoError(t, err)
	require.Equal(t, plain, got)
}
