CREATE TABLE IF NOT EXISTS blocked_users (
    blocker_username varchar(50) REFERENCES users(username) ON DELETE CASCADE,
    blocked_username varchar(50) REFERENCES users(username) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    PRIMARY KEY (blocker_username, blocked_username),
    CHECK (blocker_username <> blocked_username)
);

CREATE INDEX IF NOT EXISTS idx_blocked_users_blocker ON blocked_users (blocker_username);
CREATE INDEX IF NOT EXISTS idx_blocked_users_blocked ON blocked_users (blocked_username);
