package model

import "time"

// AttachmentInfo метаданные вложения (используется на стороне клиента, на сервере зашифрованы)
type AttachmentInfo struct {
	OriginalName string     `json:"original_name"`          // имя файла так, как его видит пользователь (без обязательной нормализации пути)
	MimeType     string     `json:"mime_type,omitempty"`    // MIME-тип содержимого (например image/png, application/pdf)
	SizeBytes    int64      `json:"size_bytes"`             // размер исходного файла в байтах до шифрования
	ModifiedAt   *time.Time `json:"modified_at,omitempty"`  // локальное время последнего изменения файла, если доступно из ФС
	SHA256       []byte     `json:"sha256,omitempty"`       // хеш исходных байт файла (например SHA-256) для проверки после скачивания и дедупликации на клиенте
	ImageWidth   *int       `json:"image_width,omitempty"`  // линейные размеры изображения в пикселях, если файл — изображение и удалось прочитать заголовок
	ImageHeight  *int       `json:"image_height,omitempty"` // линейные размеры изображения в пикселях, если файл — изображение и удалось прочитать заголовок
}

// Attachment полное вложение с телом файла (выдача по id, создание).
type Attachment struct {
	ID                string    `json:"id"`                  // ID вложения
	SecretVersionID   string    `json:"secret_version_id"`   // ID версии секрета
	DataFormatVersion int       `json:"data_format_version"` // версия формата данных
	InfoFormatVersion int       `json:"info_format_version"` // версия формата метаданных
	InfoEncrypted     []byte    `json:"info_encrypted"`      // зашифрованные метаданные
	InfoSize          int       `json:"info_size"`           // размер метаданных
	DataEncrypted     []byte    `json:"data_encrypted"`      // зашифрованные данные
	DataSize          int       `json:"data_size"`           // размер данных
	CreatedAt         time.Time `json:"created_at,omitempty"`
}

// AttachmentSummary вложение без тела (data_encrypted): списки, построение UI по зашифрованным метаданным.
type AttachmentSummary struct {
	ID                string    `json:"id"`                   // ID вложения
	SecretVersionID   string    `json:"secret_version_id"`    // ID версии секрета
	DataFormatVersion int       `json:"data_format_version"`  // версия формата данных
	InfoFormatVersion int       `json:"info_format_version"`  // версия формата метаданных
	InfoEncrypted     []byte    `json:"info_encrypted"`       // зашифрованные метаданные
	InfoSize          int       `json:"info_size"`            // размер метаданных
	DataSize          int       `json:"data_size"`            // размер данных
	CreatedAt         time.Time `json:"created_at,omitempty"` // время создания вложения
}

// AttachmentCreateRequest создание вложения (поля ciphertext заполняет клиент).
type AttachmentCreateRequest struct {
	DataFormatVersion int    `json:"data_format_version,omitempty"` // версия формата данных
	InfoFormatVersion int    `json:"info_format_version,omitempty"` // версия формата метаданных
	InfoEncrypted     []byte `json:"info_encrypted,omitempty"`      // зашифрованные метаданные
	DataEncrypted     []byte `json:"data_encrypted,omitempty"`      // зашифрованные данные
}

// AttachmentDeleteRequest удаление вложения.
type AttachmentDeleteRequest struct {
	ID string `json:"id"` // ID вложения
}
