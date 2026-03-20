ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS message_type varchar(20) DEFAULT 'text',
    ADD COLUMN IF NOT EXISTS file_url text,
    ADD COLUMN IF NOT EXISTS file_name text,
    ADD COLUMN IF NOT EXISTS mime_type varchar(255),
    ADD COLUMN IF NOT EXISTS file_size_bytes bigint,
    ADD COLUMN IF NOT EXISTS duration_seconds int;

UPDATE messages
SET message_type = 'text'
WHERE message_type IS NULL;
