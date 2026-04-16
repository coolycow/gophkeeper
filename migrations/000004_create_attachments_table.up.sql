-- Расширение pgcrypto нужно для gen_random_uuid(); на чистой БД может отсутствовать.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Таблица вложений (бинарные данные, отдельно от зашифрованного JSON в secret_versions).
-- Одна запись — один файл или иной крупный бинарный объект после шифрования на клиенте.
-- Сервер хранит только ciphertext; расшифровка и интерпретация — на клиенте.
--
-- Привязка к secret_version_id (а не к secret_id) позволяет однозначно отнести вложение
-- к снимку данных конкретной версии секрета; при удалении версии вложения удаляются каскадом.
--
-- Зашифрованные метаданные (имя файла, MIME, исходный размер и т.д.) хранятся отдельно от тела,
-- чтобы списки вложений можно было отдавать без крупного поля data_encrypted.
CREATE TABLE attachments (
    -- id вложения
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- версия секрета, к которой относится вложение
    secret_version_id UUID NOT NULL REFERENCES secret_versions(id) ON DELETE CASCADE,

    -- версия формата зашифрованного тела файла (устанавливается клиентом)
    data_format_version SMALLINT NOT NULL DEFAULT 1 CHECK (data_format_version >= 1),

    -- версия формата зашифрованного JSON/структуры в info_encrypted (устанавливается клиентом)
    info_format_version SMALLINT NOT NULL DEFAULT 1 CHECK (info_format_version >= 1),

    -- ciphertext метаданных; расшифровка только на клиенте
    info_encrypted BYTEA NOT NULL,

    -- размер info_encrypted в байтах (для квот и отображения без расшифровки)
    info_size INT NOT NULL CHECK (info_size >= 0),

    -- зашифрованные данные тела вложения; plaintext недоступен серверу
    data_encrypted BYTEA NOT NULL,

    -- размер data_encrypted в байтах (квоты, UI без скачивания тела)
    data_size INT NOT NULL CHECK (data_size >= 0),

    -- время создания записи; обновление вложений не предусмотрено — новый файл — новая строка
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP

    -- мягкое удаление записи не предусмотрено, т.к. не имеет смысла
);

-- Быстрый поиск всех вложений версии секрета
CREATE INDEX idx_attachments_secret_version_id ON attachments(secret_version_id);
