package model

// Audit — событие аудита (без чувствительных полезных данных: только идентификаторы и тип действия).
type Audit struct {
	TS              int    `json:"ts"`                          // unix timestamp события
	Action          string `json:"action"`                      // тип действия (см. пакет audit)
	UserID          string `json:"user_id,omitempty"`           // пользователь, от имени которого выполнено действие
	SecretID        string `json:"secret_id,omitempty"`         // затронутый секрет
	SecretVersionID string `json:"secret_version_id,omitempty"` // затронутая версия секрета
	AttachmentID    string `json:"attachment_id,omitempty"`     // затронутое вложение
}
