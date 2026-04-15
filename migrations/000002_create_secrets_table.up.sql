-- Создаём расширение pgcrypto для генерации UUID
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Таблица для хранения секретов
-- Поддерживается мягкое удаление записей
-- В репозитории поддерживается полное удаление записей
--
-- current_secret_version_id изначально NULL: строка secret_versions требует уже существующий secrets.id,
-- поэтому сначала вставляется секрет, затем первая версия, затем current_secret_version_id обновляется
-- в одной транзакции. Внешний ключ на secret_versions добавляется в миграции 000003.
CREATE TABLE secrets (
    -- id секрета
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- id пользователя
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- id текущей версии записи (NULL только до привязки первой версии в той же транзакции)
    current_secret_version_id UUID NULL,

    -- временные метки создания, обновления и удаления записи
    -- удалённые записи можно отображать в корзине до их полного удаления (см. репозиторий)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE NULL DEFAULT NULL
);

CREATE INDEX idx_secrets_user_id ON secrets(user_id);
CREATE UNIQUE INDEX idx_secrets_current_secret_version_id ON secrets(current_secret_version_id);
