package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/coolycow/gophkeeper/internal/logger"
	"github.com/coolycow/gophkeeper/internal/model"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Sentinel-ошибки для HardDeleteSecretVersion.
var (
	ErrSecretNotFound                   = errors.New("secret not found")
	ErrCannotDeleteCurrentSecretVersion = errors.New("cannot delete current secret version")
	ErrSecretVersionNotFound            = errors.New("secret version not found")
)

// PostgresRepository представляет репозиторий для хранения URL
type PostgresRepository struct {
	db *sql.DB
}

// checkTableExists проверяет, существует ли таблица
func (r *PostgresRepository) checkTableExists(tableName string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)`, tableName).Scan(&exists)
	return exists, err
}

// RunMigrations выполняет миграции
func (r *PostgresRepository) RunMigrations() error {
	logger.Log.Info("Running migrations")

	// Создаем экземпляр драйвера для PostgreSQL
	driver, err := pgx.WithInstance(r.db, &pgx.Config{})
	if err != nil {
		return err
	}

	// Получаем абсолютный путь к директории с миграциями
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Ищем корень проекта (где находится go.mod)
	projectRoot := wd
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}

		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			return fmt.Errorf("project root not found")
		}
		projectRoot = parent
	}

	migrationsPath := filepath.Join(projectRoot, "migrations")
	migrationsURL := "file://" + filepath.ToSlash(migrationsPath)

	// Указываем путь к директории с миграциями
	m, err := migrate.NewWithDatabaseInstance(migrationsURL, "pgx", driver)
	if err != nil {
		return err
	}

	// Применяем миграции
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

// NewPostgresRepository создает новый экземпляр URLRepository
func NewPostgresRepository(DSN string) (*PostgresRepository, error) {
	// Открываем соединение с базой данных
	db, err := sql.Open("pgx", DSN)

	// Ошибка открытия соединения
	if err != nil {
		return nil, err
	}

	// Проверяем, доступна ли база данных
	if err = db.Ping(); err != nil {
		closeErr := db.Close()
		if closeErr != nil {
			log.Printf("Error closing database: %v", closeErr)
		}
		return nil, err
	}

	repo := &PostgresRepository{db: db}

	// Проверяем, существует ли таблица users
	tableExists, err := repo.checkTableExists("users")
	if err != nil {
		return nil, err
	}

	// Если таблица users не существует, выполняем миграции, т.к. это явно первый запуск приложения на сервере
	if !tableExists {
		err = repo.RunMigrations()

		if err != nil {
			return nil, err
		}
	}

	return repo, nil
}

// //////////////////////////////////////////////////////////// МЕТОДЫ ДЛЯ РАБОТЫ С БАЗОЙ ДАННЫХ //////////////////////////////////////////////////////////////
// Close закрывает хранилище
func (r *PostgresRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// Ping проверяет доступность хранилища
func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

////////////////////////////////////////////////////////////// МЕТОДЫ ДЛЯ РАБОТЫ С ПОЛЬЗОВАТЕЛЯМИ //////////////////////////////////////////////////////////////

// GetUserByID получает пользователя по его ID
func (r *PostgresRepository) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, "select id, email, password, salt, created_at, updated_at, deleted_at from users where id = $1", userID)

	var user model.User
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Salt, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByEmail получает пользователя по его email
func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, "select id, email, password, salt, created_at, updated_at, deleted_at from users where email = $1", email)

	var user model.User
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Salt, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// CreateUser создает нового пользователя
func (r *PostgresRepository) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, `insert into users (email, password, salt) 
	values ($1, $2, $3) returning id, email, password, salt, created_at, updated_at, deleted_at`,
		user.Email, user.Password, user.Salt)

	var newUser model.User
	err := row.Scan(&newUser.ID, &newUser.Email, &newUser.Password, &newUser.Salt,
		&newUser.CreatedAt, &newUser.UpdatedAt, &newUser.DeletedAt)

	if err != nil {
		return nil, err
	}

	return &newUser, nil
}

// UpdateUser обновляет пользователя
func (r *PostgresRepository) UpdateUser(ctx context.Context, userID string, user *model.User) error {
	row := r.db.QueryRowContext(ctx, "update users set email = $1, password = $2, salt = $3, updated_at = $4, deleted_at = $5 where id = $6", user.Email, user.Password, user.Salt, user.UpdatedAt, user.DeletedAt, userID)

	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Salt, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

	if err != nil {
		return err
	}

	return nil
}

// SoftDeleteUser мягко удаляет пользователя
func (r *PostgresRepository) SoftDeleteUser(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, "update users set deleted_at = $1 where id = $2", time.Now(), userID)

	if err != nil {
		return err
	}

	return nil
}

// HardDeleteUser полностью удаляет пользователя
func (r *PostgresRepository) HardDeleteUser(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, "delete from users where id = $1", userID)

	if err != nil {
		return err
	}

	return nil
}

// GetUsersCount возвращает число записей в таблице users.
func (r *PostgresRepository) GetUsersCount(ctx context.Context) int {
	row := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users")

	var count int64
	if err := row.Scan(&count); err != nil {
		return 0
	}
	return int(count)
}

// //////////////////////////////////////////////////////////// МЕТОДЫ ДЛЯ REFRESH-ТОКЕНОВ //////////////////////////////////////////////////////////////
// CreateRefreshToken создаёт запись refresh-токена (в БД только SHA-256 от переданной клиенту строки).
func (r *PostgresRepository) CreateRefreshToken(ctx context.Context, userID string, tokenHash []byte, expiresAt time.Time) (*model.RefreshToken, error) {
	row := r.db.QueryRowContext(ctx,
		`insert into refresh_tokens (user_id, token_hash, expires_at) values ($1, $2, $3)
		returning id, user_id, token_hash, expires_at, created_at`,
		userID, tokenHash, expiresAt)

	var t model.RefreshToken
	err := row.Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// DeleteRefreshToken удаляет запись refresh-токена по первичному ключу.
func (r *PostgresRepository) DeleteRefreshToken(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `delete from refresh_tokens where id = $1`, id)
	return err
}

// FindValidRefreshTokenByHash возвращает неистёкшую запись с данным хэшем или sql.ErrNoRows.
func (r *PostgresRepository) FindValidRefreshTokenByHash(ctx context.Context, tokenHash []byte, now time.Time) (*model.RefreshToken, error) {
	row := r.db.QueryRowContext(ctx,
		`select id, user_id, token_hash, expires_at, created_at from refresh_tokens
		where token_hash = $1 and expires_at > $2`,
		tokenHash, now)

	var t model.RefreshToken
	err := row.Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

////////////////////////////////////////////////////////////// МЕТОДЫ ДЛЯ РАБОТЫ С СЕКРЕТАМИ //////////////////////////////////////////////////////////////

// GetSecretByID получает секрет по его ID
func (r *PostgresRepository) GetSecretByID(ctx context.Context, secretID string) (*model.Secret, error) {
	row := r.db.QueryRowContext(ctx, "select id, user_id, current_secret_version_id, created_at, updated_at, deleted_at from secrets where id = $1", secretID)

	var secret model.Secret
	var currentVersionID sql.NullString
	err := row.Scan(&secret.ID, &secret.UserID, &currentVersionID, &secret.CreatedAt, &secret.UpdatedAt, &secret.DeletedAt)
	if err != nil {
		return nil, err
	}
	if currentVersionID.Valid {
		secret.CurrentSecretVersionID = currentVersionID.String
	}

	return &secret, nil
}

// GetSecretsByUserID получает секреты пользователя: активные (deleted_at IS NULL), корзину (мягко удалённые) или все.
func (r *PostgresRepository) GetSecretsByUserID(ctx context.Context, userID string, scope model.SecretListScope) ([]*model.Secret, error) {
	// title_encrypted — у текущей версии (для списка на клиенте).
	q := `select s.id, s.user_id, s.current_secret_version_id, s.created_at, s.updated_at, s.deleted_at, sv.title_encrypted
from secrets s
left join secret_versions sv on sv.id = s.current_secret_version_id
where s.user_id = $1`

	// Добавляем фильтр в зависимости от scope
	switch scope {
	case model.SecretListScopeActiveOnly, model.SecretListScopeUnspecified:
		q += " and s.deleted_at is null"
	case model.SecretListScopeDeletedOnly:
		q += " and s.deleted_at is not null"
	case model.SecretListScopeAll:
		// без фильтра по deleted_at
	default:
		q += " and s.deleted_at is null"
	}

	// Добавляем сортировку по created_at
	q += " order by s.created_at asc"

	// Выполняем запрос
	rows, err := r.db.QueryContext(ctx, q, userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var secrets []*model.Secret

	for rows.Next() {
		var secret model.Secret
		var currentVersionID sql.NullString
		var titleEnc []byte
		err := rows.Scan(&secret.ID, &secret.UserID, &currentVersionID,
			&secret.CreatedAt, &secret.UpdatedAt, &secret.DeletedAt, &titleEnc)

		if err != nil {
			return nil, err
		}

		if currentVersionID.Valid {
			secret.CurrentSecretVersionID = currentVersionID.String
		}

		secret.ListTitleEncrypted = titleEnc
		secrets = append(secrets, &secret)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return secrets, nil
}

// GetSecretByUserIDAndSecretID получает секрет по его ID и ID пользователя
func (r *PostgresRepository) GetSecretByUserIDAndSecretID(ctx context.Context, userID string, secretID string) (*model.Secret, error) {
	row := r.db.QueryRowContext(ctx, `select s.id, s.user_id, s.current_secret_version_id,
	s.created_at, s.updated_at, s.deleted_at, sv.title_encrypted from secrets s
	join secret_versions sv on sv.id = s.current_secret_version_id
	where s.user_id = $1 and s.id = $2`, userID, secretID)

	var secret model.Secret

	var currentVersionID sql.NullString
	var titleEnc []byte
	err := row.Scan(&secret.ID, &secret.UserID, &currentVersionID, &secret.CreatedAt, &secret.UpdatedAt, &secret.DeletedAt, &titleEnc)

	if err != nil {
		return nil, err
	}

	if currentVersionID.Valid {
		secret.CurrentSecretVersionID = currentVersionID.String
	}
	secret.ListTitleEncrypted = titleEnc

	return &secret, nil
}

// CreateSecret создает секрет
func (r *PostgresRepository) CreateSecret(ctx context.Context, userID string, secret *model.Secret) (*model.Secret, error) {
	tx, err := r.db.BeginTx(ctx, nil)

	// Ошибка начала транзакции
	if err != nil {
		return nil, err
	}

	// 1. Секрет без current_version_id: строка в secret_versions требует уже существующий secrets.id.
	rowSecret := tx.QueryRowContext(ctx, `insert into secrets (user_id) values ($1)
		returning id, user_id, created_at, updated_at, deleted_at`, userID)

	var newSecret model.Secret
	err = rowSecret.Scan(&newSecret.ID, &newSecret.UserID, &newSecret.CreatedAt, &newSecret.UpdatedAt, &newSecret.DeletedAt)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// 2. Первая версия (version = 1).
	rowSecretVersion := tx.QueryRowContext(ctx,
		`insert into secret_versions (secret_id, version, data_format_version, data_encrypted, data_size, title_encrypted)
		values ($1, 1, $2, $3, $4, $5) returning id, secret_id, version, data_format_version, data_encrypted, data_size, created_at`,
		newSecret.ID,
		secret.SecretVersions[0].DataFormatVersion,
		secret.SecretVersions[0].DataEncrypted,
		secret.SecretVersions[0].DataSize,
		secret.SecretVersions[0].TitleEncrypted)

	var newSecretVersion model.SecretVersion
	err = rowSecretVersion.Scan(
		&newSecretVersion.ID, &newSecretVersion.SecretID, &newSecretVersion.Version,
		&newSecretVersion.DataFormatVersion, &newSecretVersion.DataEncrypted,
		&newSecretVersion.DataSize, &newSecretVersion.CreatedAt)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	newSecretVersion.TitleEncrypted = secret.SecretVersions[0].TitleEncrypted

	// 3. Привязка текущей версии (удовлетворяет FK fk_secrets_current_version_id).
	err = tx.QueryRowContext(ctx,
		`update secrets set current_secret_version_id = $1, updated_at = NOW()
		where id = $2 and user_id = $3
		returning id, user_id, current_secret_version_id, created_at, updated_at, deleted_at`,
		newSecretVersion.ID, newSecret.ID, userID,
	).Scan(&newSecret.ID, &newSecret.UserID, &newSecret.CurrentSecretVersionID, &newSecret.CreatedAt, &newSecret.UpdatedAt, &newSecret.DeletedAt)

	// Ошибка привязки секрета к версии
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// Добавляем версию в список версий секрета
	newSecret.SecretVersions = []*model.SecretVersion{&newSecretVersion}

	// Фиксируем транзакцию
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &newSecret, nil
}

// UpdateSecret обновляет секрет (TODO: доработать, т.к. фактически это неправильное обновление)
func (r *PostgresRepository) UpdateSecret(ctx context.Context, userID string, secretID string, secret *model.Secret) error {
	row := r.db.QueryRowContext(ctx, `update secrets set current_secret_version_id = $1, updated_at = $2, 
	deleted_at = $3 where user_id = $4 and id = $5
	returning id, user_id, current_secret_version_id, created_at, updated_at, deleted_at`,
		secret.CurrentSecretVersionID, secret.UpdatedAt, secret.DeletedAt, userID, secretID)

	var currentVersionID sql.NullString
	err := row.Scan(&secret.ID, &secret.UserID, &currentVersionID, &secret.CreatedAt, &secret.UpdatedAt, &secret.DeletedAt)

	// Ошибка обновления секрета
	if err != nil {
		return err
	}

	// Если ID текущей версии секрета является непустым, то обновляем его
	if currentVersionID.Valid {
		secret.CurrentSecretVersionID = currentVersionID.String
	}

	return nil
}

// SoftDeleteSecret мягко удаляет секрет (версии секрета не удаляются)
func (r *PostgresRepository) SoftDeleteSecret(ctx context.Context, userID string, secretID string) error {
	_, err := r.db.ExecContext(ctx, "update secrets set deleted_at = $1 where user_id = $2 and id = $3", time.Now(), userID, secretID)

	if err != nil {
		return err
	}

	return nil
}

// HardDeleteSecret полностью удаляет секрет (версии секрета удаляются автоматически при удалении секрета)
func (r *PostgresRepository) HardDeleteSecret(ctx context.Context, userID string, secretID string) error {
	_, err := r.db.ExecContext(ctx, "delete from secrets where user_id = $1 and id = $2", userID, secretID)

	if err != nil {
		return err
	}

	return nil
}

// CompressSecretByID удаляет все версии секрета кроме актуальной
func (r *PostgresRepository) CompressSecretByID(ctx context.Context, userID string, secretID string) error {
	_, err := r.db.ExecContext(ctx, `delete from secret_versions 
	where secret_id = $1 and secret_id in (select id from secrets where user_id = $2) 
	and version != (select current_secret_version_id from secrets where user_id = $2 and id = $1)`, secretID, userID)

	if err != nil {
		return err
	}

	return nil
}

// CompressSecretsByUserID удаляет все версии секретов кроме актуальных для всех секретов пользователя
func (r *PostgresRepository) CompressSecretsByUserID(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `delete from secret_versions 
	where secret_id in (select id from secrets where user_id = $1) 
	and version != (select current_secret_version_id from secrets where user_id = $1)`, userID)

	if err != nil {
		return err
	}

	return nil
}

// GetMaxSecretVersion получает максимальную версию секрета
func (r *PostgresRepository) GetMaxSecretVersion(ctx context.Context, userID string, secretID string) (int, error) {
	row := r.db.QueryRowContext(ctx, "select max(version) from secret_versions where secret_id = $1 and secret_id in (select id from secrets where user_id = $2)",
		secretID, userID)

	var maxVersion sql.NullInt64
	err := row.Scan(&maxVersion)

	if err != nil {
		return 0, err
	}

	if maxVersion.Valid {
		return int(maxVersion.Int64), nil
	}

	return 0, nil
}

////////////////////////////////////////////////////////////// МЕТОДЫ ДЛЯ РАБОТЫ С ИСТОРИЕЙ СЕКРЕТОВ //////////////////////////////////////////////////////////////

// GetCurrentSecretVersion получает актуальную версию секрета
func (r *PostgresRepository) GetCurrentSecretVersion(ctx context.Context, userID string, secretID string) (*model.SecretVersion, error) {
	row := r.db.QueryRowContext(ctx, `select sv.id, sv.secret_id, sv.version, sv.data_format_version, 
	sv.data_encrypted, sv.data_size, sv.created_at, sv.title_encrypted
	from secret_versions sv
	inner join secrets s on s.id = sv.secret_id and s.current_secret_version_id = sv.id
	where s.user_id = $1 and s.id = $2`, userID, secretID)

	// Получаем версию секрета
	var secretVersion model.SecretVersion
	err := row.Scan(&secretVersion.ID, &secretVersion.SecretID, &secretVersion.Version,
		&secretVersion.DataFormatVersion, &secretVersion.DataEncrypted, &secretVersion.DataSize,
		&secretVersion.CreatedAt, &secretVersion.TitleEncrypted)

	// Ошибка получения актуальной версии секрета
	if err != nil {
		return nil, err
	}

	return &secretVersion, nil
}

// GetLatestSecretVersion получает последнюю версию секрета
func (r *PostgresRepository) GetLatestSecretVersion(ctx context.Context, userID string, secretID string) (*model.SecretVersion, error) {
	// Получаем последнюю версию секрета по ID пользователя и ID секрета
	row := r.db.QueryRowContext(ctx, `select sv.id, sv.secret_id, sv.version, 
	sv.data_format_version, sv.data_encrypted, sv.data_size, sv.created_at, sv.title_encrypted
	from secret_versions as sv
	join secrets s on sv.secret_id = s.id
	where s.user_id = $1
	and sv.secret_id = $2
	and sv.version = (select max(version) from secret_versions where secret_id = $2)`, userID, secretID)

	var secretVersion model.SecretVersion
	err := row.Scan(&secretVersion.ID, &secretVersion.SecretID, &secretVersion.Version,
		&secretVersion.DataFormatVersion, &secretVersion.DataEncrypted, &secretVersion.DataSize,
		&secretVersion.CreatedAt, &secretVersion.TitleEncrypted)

	if err != nil {
		return nil, err
	}

	return &secretVersion, nil
}

// GetSecretVersionByID получает версию секрета по его ID
func (r *PostgresRepository) GetSecretVersionByID(ctx context.Context, userID string, secretID string, secretVersionID string) (*model.SecretVersion, error) {
	// Получаем версию секрета по ID пользователя, ID секрета и ID версии секрета
	row := r.db.QueryRowContext(ctx, `select sv.id, sv.secret_id, sv.version, 
	sv.data_format_version, sv.data_encrypted, sv.data_size, sv.created_at, sv.title_encrypted
	from secret_versions as sv
	join secrets s on sv.secret_id = s.id
	where s.user_id = $1
	and sv.secret_id = $2
	and sv.id = $3`, userID, secretID, secretVersionID)

	var secretVersion model.SecretVersion
	err := row.Scan(&secretVersion.ID, &secretVersion.SecretID, &secretVersion.Version,
		&secretVersion.DataFormatVersion, &secretVersion.DataEncrypted, &secretVersion.DataSize,
		&secretVersion.CreatedAt, &secretVersion.TitleEncrypted)

	if err != nil {
		return nil, err
	}

	return &secretVersion, nil
}

// GetAllSecretHistories получает все версии секрета
func (r *PostgresRepository) GetAllSecretHistories(ctx context.Context, userID string, secretID string) ([]*model.SecretVersion, error) {
	// Получаем все версии секрета по ID пользователя и ID секрета
	rows, err := r.db.QueryContext(ctx, `select sv.id, sv.secret_id, sv.version, 
	sv.data_format_version, sv.data_encrypted, sv.data_size, sv.created_at, sv.title_encrypted
	from secret_versions as sv
	join secrets s on sv.secret_id = s.id
	where s.user_id = $1
	and sv.secret_id = $2`, userID, secretID)

	// Ошибка получения всех версий секрета
	if err != nil {
		return nil, err
	}

	// Закрываем строки
	defer rows.Close()

	// Создаем слайс версий секрета
	var secretHistories []*model.SecretVersion

	// Цикл по строкам
	for rows.Next() {
		var secretVersion model.SecretVersion
		err := rows.Scan(&secretVersion.ID, &secretVersion.SecretID, &secretVersion.Version,
			&secretVersion.DataFormatVersion, &secretVersion.DataEncrypted, &secretVersion.DataSize,
			&secretVersion.CreatedAt, &secretVersion.TitleEncrypted)

		if err != nil {
			return nil, err
		}

		secretHistories = append(secretHistories, &secretVersion)
	}

	// Ошибка получения всех версий секрета
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return secretHistories, nil
}

// CreateSecretVersion создает версию секрета
func (r *PostgresRepository) CreateSecretVersion(ctx context.Context, userID string, secretID string, secretVersion *model.SecretVersion) (*model.SecretVersion, error) {
	// Начинаем транзакцию, т.к. возможна конкуренция при создании версии секрета в разных потоках
	tx, err := r.db.BeginTx(ctx, nil)

	if err != nil {
		return nil, err
	}

	// Блокируем строку секрета для обновления
	row := tx.QueryRowContext(ctx, "select 1 from secrets where user_id = $1 and id = $2 for update", userID, secretID)

	// Получаем dummy
	var dummy int
	err = row.Scan(&dummy)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// Получаем максимальную версию секрета
	rowMaxVersion := tx.QueryRowContext(ctx, "select coalesce(max(version), 0) from secret_versions where secret_id = $1", secretID)

	var maxVersion int
	err = rowMaxVersion.Scan(&maxVersion)

	// Ошибка получения максимальной версии секрета
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// Увеличиваем версию секрета на 1
	secretVersion.Version = maxVersion + 1

	// Создаем версию секрета по ID пользователя, ID секрета и версии секрета
	row = tx.QueryRowContext(ctx, `insert into secret_versions 
	(secret_id, version, data_format_version, data_encrypted, data_size, title_encrypted) 
	values ($1, $2, $3, $4, $5, $6) 
	returning id, secret_id, version, data_format_version, data_encrypted, data_size, created_at`,
		secretID, secretVersion.Version, secretVersion.DataFormatVersion, secretVersion.DataEncrypted, secretVersion.DataSize, secretVersion.TitleEncrypted)

	// Получаем версию секрета
	var newSecretVersion model.SecretVersion
	err = row.Scan(&newSecretVersion.ID, &newSecretVersion.SecretID, &newSecretVersion.Version, &newSecretVersion.DataFormatVersion,
		&newSecretVersion.DataEncrypted, &newSecretVersion.DataSize, &newSecretVersion.CreatedAt)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	newSecretVersion.TitleEncrypted = secretVersion.TitleEncrypted

	// Привязываем версию секрета к секрету как текущую версию
	_, err = tx.ExecContext(ctx, `update secrets 
	set current_secret_version_id = $1, updated_at = NOW() 
	where user_id = $2 and id = $3`, newSecretVersion.ID, userID, secretID)

	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// Фиксируем транзакцию
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &newSecretVersion, nil
}

// HardDeleteSecretVersion полностью удаляет версию секрета (мягкое удаление невозможно, т.к. не имеет смысла)
func (r *PostgresRepository) HardDeleteSecretVersion(ctx context.Context, userID string, secretID string, secretVersionID string) error {
	// Удаляем версию секрета
	_, err := r.db.ExecContext(ctx, `
		delete from secret_versions sv
		using secrets s
		where sv.secret_id = s.id
		  and s.user_id = $1
		  and s.id = $2
		  and sv.id = $3`,
		userID, secretID, secretVersionID,
	)

	// Ошибка удаления версии секрета
	if err != nil {
		return err
	}

	return nil
}

// HardDeleteOldestSecretVersion полностью удаляет самую старую версию секрета (должна быть не актуальной версией)
// Логика: удаляем ближайшую к текущей не актуальную версию (проверка по минимуму версии исключая текущую).
// Если у секрета всего одна версия, то удаляем её и по сути секрет становится пустым.
func (r *PostgresRepository) HardDeleteOldestSecretVersion(ctx context.Context, userID string, secretID string) error {
	// Получаем количество версий секрета
	row := r.db.QueryRowContext(ctx, `select count(*) from secret_versions sv
	join secrets s on sv.secret_id = s.id
	where s.user_id = $1 and sv.secret_id = $2`, userID, secretID)

	var count int
	err := row.Scan(&count)

	if err != nil {
		return err
	}

	// Если количество версий секрета равно 1, то удаляем её и по сути секрет становится пустым
	if count == 1 {
		_, err := r.db.ExecContext(ctx, `delete from secret_versions sv
		using secrets s
		where sv.secret_id = s.id
		  and s.user_id = $1
		  and sv.secret_id = $2`, secretID, userID)

		if err != nil {
			return err
		}

		return nil
	}

	// Удаляем версию секрета с минимальной версией (проверка по минимуму версии исключая текущую)
	_, err = r.db.ExecContext(ctx, `delete from secret_versions 
	where secret_id = $1 
	and secret_id in (select id from secrets where user_id = $2) 
	and version = (
		select min(version) from secret_versions where secret_id = $1
		and secret_id in (select id from secrets where user_id = $2)
		and version <> (select current_secret_version_id from secrets where user_id = $2 and id = $3))`, secretID, userID, secretID)

	if err != nil {
		return err
	}

	return nil
}

// RestoreSecretVersion восстанавливает версию секрета (делает её текущей для секрета)
func (r *PostgresRepository) RestoreSecretVersion(ctx context.Context, userID string, secretID string, secretVersionID string) error {
	_, err := r.db.ExecContext(ctx, `update secrets 
	set current_secret_version_id = $1, updated_at = NOW() 
	where user_id = $2 and id = $3`, secretVersionID, userID, secretID)

	if err != nil {
		return err
	}

	return nil
}

////////////////////////////////////////////////////////////// МЕТОДЫ ДЛЯ РАБОТЫ С ВЛОЖЕНИЯМИ //////////////////////////////////////////////////////////////

// GetAttachmentByID получает вложение по его ID
func (r *PostgresRepository) GetAttachmentByID(ctx context.Context, userID string, attachmentID string) (*model.Attachment, error) {
	// Получаем вложение по его ID по ID пользователя и ID вложения
	row := r.db.QueryRowContext(ctx, `select a.id, a.secret_version_id, a.data_format_version, 
	a.info_format_version, a.info_encrypted, a.info_size, a.data_encrypted, a.data_size, a.created_at 
	from attachments as a 
	where a.id = $1 and a.secret_version_id in (
		select sv.id from secret_versions as sv where sv.secret_id in (select s.id from secrets as s where s.user_id = $2)
	)`, attachmentID, userID)

	var attachment model.Attachment
	err := row.Scan(&attachment.ID, &attachment.SecretVersionID, &attachment.DataFormatVersion, &attachment.InfoFormatVersion,
		&attachment.InfoEncrypted, &attachment.InfoSize, &attachment.DataEncrypted, &attachment.DataSize, &attachment.CreatedAt)

	// Ошибка получения вложения по его ID
	if err != nil {
		return nil, err
	}

	return &attachment, nil
}

// GetAttachmentsBySecretID получает все вложения по ID секрета (без data_encrypted — только метаданные для списка)
func (r *PostgresRepository) GetAttachmentsBySecretID(ctx context.Context, userID string, secretID string) ([]*model.AttachmentSummary, error) {
	// Получаем все вложения по ID секрета по ID пользователя и ID секрета
	rows, err := r.db.QueryContext(ctx, `select a.id, a.secret_version_id, a.data_format_version, 
	a.info_format_version, a.info_encrypted, a.info_size, a.data_encrypted, a.data_size, a.created_at 
	from attachments as a
	where a.secret_version_id in (
		select sv.id from secret_versions as sv where sv.secret_id in (select s.id from secrets as s where s.user_id = $2)
	) and sv.secret_id = $1`, userID, secretID)

	// Ошибка получения всех вложений по ID секрета
	if err != nil {
		return nil, err
	}

	// Закрываем строки
	defer rows.Close()

	// Создаем слайс вложений
	var attachments []*model.AttachmentSummary

	// Цикл по строкам
	for rows.Next() {
		var summary model.AttachmentSummary
		err := rows.Scan(&summary.ID, &summary.SecretVersionID, &summary.DataFormatVersion, &summary.InfoFormatVersion,
			&summary.InfoEncrypted, &summary.InfoSize, &summary.DataSize, &summary.CreatedAt)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, &summary)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return attachments, nil
}

// GetAttachmentsBySecretVersionID получает все вложения по ID версии секрета (без data_encrypted — только метаданные для списка)
func (r *PostgresRepository) GetAttachmentsBySecretVersionID(ctx context.Context, userID string, secretVersionID string) ([]*model.AttachmentSummary, error) {
	// Получаем все вложения по ID версии секрета по ID пользователя и ID версии секрета
	rows, err := r.db.QueryContext(ctx, `select a.id, a.secret_version_id, a.data_format_version, 
	a.info_format_version, a.info_encrypted, a.info_size, a.data_encrypted, a.data_size, a.created_at 
	from attachments as a
	where secret_version_id = $1 
	and secret_version_id in (
		select sv.id from secret_versions as sv where sv.secret_id in (select s.id from secrets as s where s.user_id = $2)
	)`, secretVersionID, userID)

	// Ошибка получения всех вложений по ID версии секрета
	if err != nil {
		return nil, err
	}

	// Закрываем строки
	defer rows.Close()

	// Создаем слайс вложений
	var attachments []*model.AttachmentSummary

	// Цикл по строкам
	for rows.Next() {
		var summary model.AttachmentSummary
		err := rows.Scan(&summary.ID, &summary.SecretVersionID, &summary.DataFormatVersion, &summary.InfoFormatVersion,
			&summary.InfoEncrypted, &summary.InfoSize, &summary.DataSize, &summary.CreatedAt)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, &summary)
	}

	// Ошибка получения всех вложений по ID версии секрета
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return attachments, nil
}

// CreateAttachment создает вложение
func (r *PostgresRepository) CreateAttachment(ctx context.Context, userID string, secretVersionID string, attachment *model.Attachment) (*model.Attachment, error) {
	row := r.db.QueryRowContext(ctx, `insert into attachments 
	(secret_version_id, data_format_version, info_format_version, info_encrypted, info_size, data_encrypted, data_size)
	values ($1, $2, $3, $4, $5, $6, $7) 
	returning id, secret_version_id, data_format_version, info_format_version, info_encrypted, 
	info_size, data_encrypted, data_size, created_at`,
		secretVersionID, attachment.DataFormatVersion, attachment.InfoFormatVersion, attachment.InfoEncrypted, attachment.InfoSize, attachment.DataEncrypted, attachment.DataSize)

	var newAttachment model.Attachment
	err := row.Scan(&newAttachment.ID, &newAttachment.SecretVersionID, &newAttachment.DataFormatVersion, &newAttachment.InfoFormatVersion,
		&newAttachment.InfoEncrypted, &newAttachment.InfoSize, &newAttachment.DataEncrypted, &newAttachment.DataSize, &newAttachment.CreatedAt)

	// Ошибка создания вложения
	if err != nil {
		return nil, err
	}

	return &newAttachment, nil
}

// DeleteAttachment удаляет вложение
func (r *PostgresRepository) HardDeleteAttachment(ctx context.Context, userID string, attachmentID string) error {
	_, err := r.db.ExecContext(ctx, `delete from attachments 
	where id = $1 
	and secret_version_id in 
	(select id from secret_versions where secret_id in (select id from secrets where user_id = $2))`, attachmentID, userID)

	// Ошибка удаления вложения
	if err != nil {
		return err
	}

	return nil
}
