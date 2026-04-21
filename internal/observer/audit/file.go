package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/coolycow/gophkeeper/internal/model"
)

// FileReceiver записывает события аудита в файл (каждое событие — новая строка).
type FileReceiver struct {
	path string
	mu   sync.Mutex
}

// NewFileReceiver создаёт приёмник в файл.
func NewFileReceiver(path string) *FileReceiver {
	return &FileReceiver{path: path}
}

// Send добавляет событие в конец файла в виде одной строки JSON.
func (f *FileReceiver) Send(event *model.Audit) error {
	// Преобразуем событие в JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	// Добавляем символ новой строки
	data = append(data, '\n')

	// Открываем файл для записи
	file, err := os.OpenFile(f.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open audit file %s: %w", f.path, err)
	}
	defer func() { _ = file.Close() }()

	// Записываем событие в файл
	f.mu.Lock()
	_, err = file.Write(data)
	f.mu.Unlock()

	// Если ошибка при записи в файл, возвращаем ошибку
	if err != nil {
		return fmt.Errorf("write audit event to file: %w", err)
	}

	return nil
}
