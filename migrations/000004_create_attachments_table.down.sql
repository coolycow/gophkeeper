-- Удаляем индекс по версии секрета
DROP INDEX IF EXISTS idx_attachments_secret_version_id;

-- Удаляем таблицу вложений
DROP TABLE IF EXISTS attachments CASCADE;
