package model

import "time"

// Secret модель секрета
type Secret struct {
	ID                     string           `json:"id"`                        // ID секрета
	UserID                 string           `json:"user_id"`                   // ID пользователя
	CurrentSecretVersionID string           `json:"current_secret_version_id"` // ID текущей версии секрета
	CreatedAt              *time.Time       `json:"created_at,omitempty"`      // время создания секрета
	UpdatedAt              *time.Time       `json:"updated_at,omitempty"`      // время обновления секрета
	DeletedAt              *time.Time       `json:"deleted_at,omitempty"`      // время удаления секрета
	SecretVersions         []*SecretVersion `json:"secret_versions,omitempty"` // версии секрета
	// ListTitleEncrypted — title_encrypted текущей версии, подставляется в выборке GetSecretsByUserID (с ним же список в API).
	ListTitleEncrypted []byte `json:"list_title_encrypted,omitempty"`
}

// SecretCreateRequest модель создания секрета
type SecretCreateRequest struct {
	DataEncrypted     []byte `json:"encrypted_data,omitempty"`      // зашифрованные данные
	TitleEncrypted    []byte `json:"title_encrypted,omitempty"`     // зашифрованное отображаемое имя (текущая версия)
	DataFormatVersion int    `json:"data_format_version,omitempty"` // версия формата (gRPC / API)
}

// SecretUpdateRequest модель обновления секрета
type SecretUpdateRequest struct {
	DataEncrypted  []byte `json:"binary_data,omitempty"` // зашифрованные данные
	TitleEncrypted []byte `json:"title_encrypted,omitempty"`
}

// SecretDeleteRequest модель удаления секрета
type SecretDeleteRequest struct {
	ID string `json:"id"`
}
