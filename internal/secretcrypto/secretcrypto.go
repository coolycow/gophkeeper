// Package secretcrypto шифрует полезную нагрузку секрета на клиенте до отправки на сервер.
// Формат: префикс GK01 + nonce + AES-256-GCM; ключ — Argon2id от пароля и соли пользователя (hex).
package secretcrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/argon2"
)

// DataFormatVersion отправляется в gRPC как data_format_version вместе с ciphertext.
const DataFormatVersion = 1

const (
	magic    = "GK01" // magic число для проверки, что данные зашифрованы правильно
	magicLen = 4      // длина magic числа
	nonceLen = 12     // длина nonce
	keyLen   = 32     // длина ключа

	// Параметры Argon2id (ориентир OWASP для парольных схем).
	argon2Time      uint32 = 2         // количество времени
	argon2MemoryKiB uint32 = 64 * 1024 // 64 MiB памяти
	argon2Threads   uint8  = 4         // количество потоков
)

// ErrDecrypt ошибка расшифровки
var ErrDecrypt = errors.New("secretcrypto: decrypt failed (wrong password or corrupted data)")

// decodeSalt декодирует hex строку в байты
func decodeSalt(saltHex string) ([]byte, error) {
	// Убираем пробелы и переводы строки
	saltHex = strings.TrimSpace(saltHex)

	// Проверяем, что соль не пустая
	if saltHex == "" {
		return nil, fmt.Errorf("empty salt")
	}

	// Декодируем hex строку в байты
	b, err := hex.DecodeString(saltHex)
	if err != nil {
		return nil, fmt.Errorf("salt hex: %w", err)
	}

	return b, nil
}

// deriveKey генерирует ключ из пароля и соли с помощью Argon2id
func deriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, argon2Time, argon2MemoryKiB, argon2Threads, keyLen)
}

// sealAESGCM шифрует plaintext с помощью AES-256-GCM
func sealAESGCM(key, nonce, plaintext []byte) ([]byte, error) {
	// Создаем новый блок AES
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Создаем новый блок GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm.Seal(nil, nonce, plaintext, nil), nil
}

// openAESGCM расшифровывает ciphertext с помощью AES-256-GCM
func openAESGCM(key, nonce, ciphertext []byte) ([]byte, error) {
	// Создаем новый блок AES
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Создаем новый блок GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, nonce, ciphertext, nil)
}

// Encrypt шифрует plaintext (AES-256-GCM), ключ — Argon2id от пароля и соли пользователя (hex).
func Encrypt(plaintext []byte, password string, saltHex string) ([]byte, error) {
	// Декодируем соль
	salt, err := decodeSalt(saltHex)
	if err != nil {
		return nil, err
	}
	key := deriveKey(password, salt)

	// Генерируем nonce
	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Шифруем plaintext
	sealed, err := sealAESGCM(key, nonce, plaintext)
	if err != nil {
		return nil, err
	}

	out := make([]byte, 0, magicLen+nonceLen+len(sealed))
	out = append(out, []byte(magic)...)
	out = append(out, nonce...)
	out = append(out, sealed...)
	return out, nil
}

// Decrypt расшифровывает блоб, полученный из Encrypt.
func Decrypt(blob []byte, password string, saltHex string) ([]byte, error) {
	// Проверяем, что длина blob не меньше magicLen+nonceLen
	if len(blob) < magicLen+nonceLen {
		return nil, ErrDecrypt
	}

	// Проверяем, что magic число совпадает
	if string(blob[:magicLen]) != magic {
		return nil, ErrDecrypt
	}

	// Декодируем соль
	salt, err := decodeSalt(saltHex)
	if err != nil {
		return nil, err
	}

	// Генерируем ключ
	key := deriveKey(password, salt)
	nonce := blob[magicLen : magicLen+nonceLen]
	ct := blob[magicLen+nonceLen:]

	// Расшифровываем ciphertext
	plain, err := openAESGCM(key, nonce, ct)

	if err != nil {
		return nil, ErrDecrypt
	}

	return plain, nil
}
