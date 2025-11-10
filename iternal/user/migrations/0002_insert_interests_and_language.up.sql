INSERT INTO languages (lang_name) VALUES
    ('en'),
    ('ru')
ON CONFLICT (lang_name) DO NOTHING;

INSERT INTO interests (interest_name) VALUES
    ('music'),
    ('movies'),
    ('sport'),
    ('books')
ON CONFLICT (interest_name) DO NOTHING;