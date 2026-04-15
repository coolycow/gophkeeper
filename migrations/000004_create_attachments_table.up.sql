-- Расширение pgcrypto нужно для gen_random_uuid(); на чистой БД может отсутствовать.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Таблица вложений (бинарные данные, отдельно от зашифрованного JSON в secret_versions).
-- Одна запись — один файл или иной крупный бинарный объект после шифрования на клиенте.
-- Сервер хранит только ciphertext; расшифровка и интерпретация — на клиенте.
--
-- Привязка к secret_version_id (а не к secret_id) позволяет однозначно отнести вложение
-- к снимку данных конкретной версии секрета; при удалении версии вложения удаляются каскадом.
CREATE TABLE attachments (
    -- id вложения
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- версия секрета, к которой относится вложение
    secret_version_id UUID NOT NULL REFERENCES secret_versions(id) ON DELETE CASCADE,

    -- версия формата зашифрованной полезной нагрузки вложения (метаданные файла, обёртка ключа и т.д.)
    -- согласуется по смыслу с data_format_version в secret_versions, но независима по номеру
    data_format_version SMALLINT NOT NULL DEFAULT 1 CHECK (data_format_version >= 1),

    -- зашифрованные данные вложения; plaintext недоступен серверу
    data_encrypted BYTEA NOT NULL,

    -- размер полезной нагрузки после шифрования (или согласованный с клиентом смысл), для квот и UI
    data_size INT NOT NULL CHECK (data_size >= 0),

    -- время создания записи; обновление вложений не предусмотрено — новый файл — новая строка
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP

    -- мягкое удаление записи не предусмотрено, т.к. не имеет смысла
);

-- Быстрый поиск всех вложений версии секрета
CREATE INDEX idx_attachments_secret_version_id ON attachments(secret_version_id);
