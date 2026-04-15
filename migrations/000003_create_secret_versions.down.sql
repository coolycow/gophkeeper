-- Удаляем внешний ключ на secret_versions
ALTER TABLE secrets DROP CONSTRAINT IF EXISTS fk_secrets_current_version_id;

-- Удаляем индексы на secret_versions
DROP INDEX IF EXISTS idx_secret_versions_secret_id;
DROP INDEX IF EXISTS idx_secret_versions_secret_id_version;

-- Удаляем таблицу secret_versions
DROP TABLE IF EXISTS secret_versions CASCADE;