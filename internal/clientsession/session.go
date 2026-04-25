// Package clientsession сохраняет refresh токен и идентификатор сервера между запусками (0600 файл).
package clientsession

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// File хранится на диске как JSON.
type File struct {
	ServerAddress string `json:"server_address"`
	Email         string `json:"email"`
	RefreshToken  string `json:"refresh_token"`
}

// DefaultDir возвращает OS config поддиректорию для GophKeeper (создает ее при Save).
func DefaultDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "gophkeeper"), nil
}

// DefaultPath это DefaultDir + session.json.
func DefaultPath() (string, error) {
	dir, err := DefaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "session.json"), nil
}

// Save записывает сессию атомарно с режимом 0600.
func Save(path string, f *File) error {
	if f == nil {
		return errors.New("clientsession: nil file")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// Load читает сессию или возвращает nil файл, если отсутствует.
func Load(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

// Clear удаляет файл сессии, если он присутствует.
func Clear(path string) error {
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
