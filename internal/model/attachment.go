package model

import "time"

// Attachment модель вложения
type Attachment struct {
	ID                string    `json:"id"`                   // ID вложения
	SecretVersionID   string    `json:"secret_version_id"`    // ID версии секрета
	DataFormatVersion int       `json:"data_format_version"`  // версия формата данных
	DataEncrypted     []byte    `json:"data_encrypted"`       // зашифрованные данные
	DataSize          int       `json:"data_size"`            // размер данных
	CreatedAt         time.Time `json:"created_at,omitempty"` // время создания вложения
}

// AttachmentCreateRequest модель создания вложения
type AttachmentCreateRequest struct {
	DataEncrypted []byte `json:"data_encrypted,omitempty"` // зашифрованные данные
}

// AttachmentDeleteRequest модель удаления вложения
type AttachmentDeleteRequest struct {
	ID string `json:"id"`
}
