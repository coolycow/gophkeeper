package model

import "time"

// SecretVersion модель версии секрета
type SecretVersion struct {
	ID                string    `json:"id"`                   // ID версии секрета
	SecretID          string    `json:"secret_id"`            // ID секрета
	Version           int       `json:"version"`              // версия секрета
	DataFormatVersion int       `json:"data_format_version"`  // версия формата данных
	DataEncrypted     []byte    `json:"data_encrypted"`       // зашифрованные данные
	DataSize          int       `json:"data_size"`            // размер данных
	CreatedAt         time.Time `json:"created_at,omitempty"` // время создания версии секрета
	UpdatedAt         time.Time `json:"updated_at,omitempty"` // время обновления версии секрета
	DeletedAt         time.Time `json:"deleted_at,omitempty"` // время удаления версии секрета
}

// GetDataSize возвращает размер данных
func (s *SecretVersion) GetDataSize() int {
	return len(s.DataEncrypted)
}
