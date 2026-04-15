-- Создаём расширение pgcrypto для генерации UUID
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Таблица для хранения версий секретов
-- Обновление записей не допускается
CREATE TABLE secret_versions (
    -- id версии записи
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- id секрета
    secret_id UUID NOT NULL REFERENCES secrets(id) ON DELETE CASCADE,

    -- версия записи
    -- стандартная сортировка по версии записи по убыванию на уровне репозитория
    -- общее количество версий записи может быть ограничено конфигурацией сервера
    version INT NOT NULL CHECK (version >= 1),

    -- версия формата данных
    -- версия формата данных может быть увеличена, но не может быть уменьшена
    -- если на клиенте поддерживается большая версия формата данных, чем на сервере, то клиент конвертирует данные в новый формат
    -- если на клиенте поддерживается меньшая версия формата данных, чем на сервере, то клиент не сможет прочитать данные
    -- гарантируется обратная совместимость с предыдущими версиями формата данных
    data_format_version SMALLINT NOT NULL DEFAULT 1 CHECK (data_format_version >= 1),

    -- зашифрованные данные записи, расшифровать может только клиент
    -- внутри зашифрованные данные хранятся в формате JSON, который содержит структуру соответствующую версии формата данных
    data_encrypted BYTEA NOT NULL,

    -- размер данных записи, не может быть отрицательным, но может быть равен нулю, нужен для быстрой статистики
    data_size INT NOT NULL CHECK (data_size >= 0),

    -- временная метка создания записи
    -- временные метки обновления и удаления записи не используются
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- индексы для быстрого поиска версий секрета
CREATE INDEX idx_secret_versions_secret_id ON secret_versions(secret_id);
-- уникальный индекс для быстрого поиска версии секрета и невозможности создания двух версий с одной и той же версией
CREATE UNIQUE INDEX idx_secret_versions_secret_id_version ON secret_versions(secret_id, version);

-- Связь secrets → secret_versions возможна только после создания обеих таблиц.
-- Удаление строки текущей версии каскадно удаляет секрет (ссылка «текущая версия» исчезла).
ALTER TABLE secrets
    ADD CONSTRAINT fk_secrets_current_secret_version_id
    FOREIGN KEY (current_secret_version_id) REFERENCES secret_versions(id) ON DELETE CASCADE;