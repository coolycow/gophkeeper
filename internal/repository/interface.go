// Package repository определяет интерфейсы и реализации хранилища URL.
package repository

import (
	"context"

	"github.com/coolycow/gophkeeper/internal/model"
)

// GophKeeperRepository — интерфейс хранилища: сохранение/получение секретов, пользователи, пинг, миграции.
type GophKeeperRepository interface {
	// Методы для работы с базой данных
	Close() error                   // закрывает соединение с базой данных
	Ping(ctx context.Context) error // проверяет доступность базы данных
	RunMigrations() error           // выполняет миграции

	////////////////////////////////////////////////////////////// МЕТОДЫ ДЛЯ РАБОТЫ С ПОЛЬЗОВАТЕЛЯМИ //////////////////////////////////////////////////////////////
	GetUserByID(ctx context.Context, userID string) (*model.User, error)   // получает пользователя по его ID
	GetUserByEmail(ctx context.Context, email string) (*model.User, error) // получает пользователя по его email
	CreateUser(ctx context.Context, user *model.User) (*model.User, error) // создает нового пользователя
	UpdateUser(ctx context.Context, userID string, user *model.User) error // обновляет пользователя
	SoftDeleteUser(ctx context.Context, userID string) error               // мягко удаляет пользователя
	HardDeleteUser(ctx context.Context, userID string) error               // полностью удаляет пользователя
	GetUsersCount(ctx context.Context) int                                 // возвращает количество пользователей

	////////////////////////////////////////////////////////////// МЕТОДЫ ДЛЯ РАБОТЫ С СЕКРЕТАМИ //////////////////////////////////////////////////////////////
	GetSecretByID(ctx context.Context, secretID string) (*model.Secret, error)                               // получает секрет по его ID
	GetSecretsByUserID(ctx context.Context, userID string) ([]*model.Secret, error)                          // получает все секреты пользователя
	GetSecretByUserIDAndSecretID(ctx context.Context, userID string, secretID string) (*model.Secret, error) // получает секрет по его ID и ID пользователя
	CreateSecret(ctx context.Context, userID string, secret *model.Secret) (*model.Secret, error)            // создает секрет
	UpdateSecret(ctx context.Context, userID string, secretID string, secret *model.Secret) error            // обновляет секрет
	SoftDeleteSecret(ctx context.Context, userID string, secretID string) error                              // мягко удаляет секрет
	HardDeleteSecret(ctx context.Context, userID string, secretID string) error                              // полностью удаляет секрет
	CompressSecretByID(ctx context.Context, userID string, secretID string) error                            // удаляет все версии секрета кроме актуальной
	CompressSecretsByUserID(ctx context.Context, userID string) error                                        // удаляет все версии секретов кроме актуальной для всех секретов пользователя
	GetMaxSecretVersion(ctx context.Context, userID string, secretID string) (int, error)                    // получает максимальную версию секрета

	////////////////////////////////////////////////////////////// МЕТОДЫ ДЛЯ РАБОТЫ С ИСТОРИЕЙ СЕКРЕТОВ //////////////////////////////////////////////////////////////
	GetCurrentSecretVersion(ctx context.Context, userID string, secretID string) (*model.SecretVersion, error)                                 // получает текущую версию секрета
	GetLatestSecretVersion(ctx context.Context, userID string, secretID string) (*model.SecretVersion, error)                                  // получает последнюю версию секрета
	GetSecretVersionByID(ctx context.Context, userID string, secretID string, secretVersionID string) (*model.SecretVersion, error)            // получает версию секрета по его ID
	GetAllSecretHistories(ctx context.Context, userID string, secretID string) ([]*model.SecretVersion, error)                                 // получает все версии секрета
	CreateSecretVersion(ctx context.Context, userID string, secretID string, secretVersion *model.SecretVersion) (*model.SecretVersion, error) // создает версию секрета
	HardDeleteSecretVersion(ctx context.Context, userID string, secretID string, secretVersionID string) error                                 // полностью удаляет версию секрета (мягкое удаление невозможно, т.к. не имеет смысла)
	RestoreSecretVersion(ctx context.Context, userID string, secretID string, secretVersionID string) error                                    // восстанавливает версию секрета (делает её текущей для секрета)

	////////////////////////////////////////////////////////////// МЕТОДЫ ДЛЯ РАБОТЫ С ВЛОЖЕНИЯМИ //////////////////////////////////////////////////////////////
	GetAttachmentByID(ctx context.Context, userID string, attachmentID string) (*model.Attachment, error)                                 // получает вложение по его ID
	GetAttachmentsBySecretID(ctx context.Context, userID string, secretID string) ([]*model.AttachmentSummary, error)                     // получает все вложения по ID секрета (без тела data_encrypted)
	GetAttachmentsBySecretVersionID(ctx context.Context, userID string, secretVersionID string) ([]*model.AttachmentSummary, error)       // получает все вложения по ID версии секрета (без тела data_encrypted)
	CreateAttachment(ctx context.Context, userID string, secretVersionID string, attachment *model.Attachment) (*model.Attachment, error) // создает вложение
	HardDeleteAttachment(ctx context.Context, userID string, attachmentID string) error                                                   // удаляет вложение
}
