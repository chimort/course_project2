ALTER TABLE chat_histories
    ADD COLUMN IF NOT EXISTS search_mode varchar(50);

UPDATE chat_histories
SET search_mode = 'MATCH_MODE_DEFAULT'
WHERE search_mode IS NULL;
