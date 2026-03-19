INSERT INTO languages (lang_name) VALUES
    ('en'),
    ('ru')
ON CONFLICT (lang_name) DO NOTHING;

INSERT INTO interests (interest_name) VALUES
    ('music'),
    ('movies'),
    ('sport'),
    ('books'),
    ('art'),
    ('painting'),
    ('photography'),
    ('design'),
    ('theatre'),
    ('dance'),
    ('travel'),
    ('hiking'),
    ('cooking'),
    ('gaming'),
    ('anime'),
    ('science'),
    ('technology'),
    ('programming'),
    ('history'),
    ('philosophy'),
    ('fitness'),
    ('yoga'),
    ('architecture')
ON CONFLICT (interest_name) DO NOTHING;
