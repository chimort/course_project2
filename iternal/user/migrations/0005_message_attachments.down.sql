ALTER TABLE messages
    DROP COLUMN IF EXISTS duration_seconds,
    DROP COLUMN IF EXISTS file_size_bytes,
    DROP COLUMN IF EXISTS mime_type,
    DROP COLUMN IF EXISTS file_name,
    DROP COLUMN IF EXISTS file_url,
    DROP COLUMN IF EXISTS message_type;
