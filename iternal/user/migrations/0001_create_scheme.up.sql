create table users (
    username varchar(50) PRIMARY KEY,
    email varchar(100) unique not null,
    password_hash text not null,
    age int,
    gender varchar(10)
);

create table languages (
    id serial PRIMARY KEY,
    lang_name varchar(100) unique not null
);

create table user_languages (
    username varchar(50) REFERENCES users(username) on delete cascade,
    language_id int REFERENCES languages(id),
    proficiency_level varchar(20),
    PRIMARY KEY (username, language_id)
);

create table interests(
    id serial PRIMARY KEY,
    interest_name varchar(100) unique not null
);

create table user_interests(
    username varchar(50) REFERENCES users(username) on delete cascade,
    interest_id int REFERENCES interests(id),
    PRIMARY KEY (username, interest_id)
);

create table chats (
    id serial PRIMARY KEY,
    chat_type varchar(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

create table chat_participants (
    chat_id int REFERENCES chats(id) on delete cascade,
    username varchar(50) REFERENCES users(username) on delete cascade,
    last_read_at TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY(chat_id, username)
);

create table chat_histories(
    chat_id INT PRIMARY KEY REFERENCES chats(id) on delete cascade,
    last_message_at TIMESTAMP WITH TIME ZONE,
    match_score float,
    match_tags TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

create table messages(
    id BIGSERIAL PRIMARY KEY,
    chat_id int REFERENCES chats(id) on delete cascade,
    sender_name varchar(50) REFERENCES users(username) on delete set null,
    content text not null,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);


ALTER TABLE user_languages
    ADD CONSTRAINT fk_user_languages_user FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE;

ALTER TABLE user_interests
    ADD CONSTRAINT fk_user_interests_user FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE;

ALTER TABLE chat_participants
    ADD CONSTRAINT fk_chat_participants_chat FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_chat_participants_user FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE;

ALTER TABLE messages
    ADD CONSTRAINT fk_messages_chat FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_messages_sender FOREIGN KEY (sender_name) REFERENCES users(username) ON DELETE SET NULL;

ALTER TABLE chat_histories
    ADD CONSTRAINT fk_chat_histories_chat FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE;


CREATE INDEX idx_chat_participant_user ON chat_participants (username, chat_id);
CREATE INDEX idx_messages_chat_created ON messages (chat_id, created_at DESC);
CREATE INDEX idx_chat_histories_lastmsg ON chat_histories (last_message_at DESC);
CREATE INDEX idx_user_languages_language ON user_languages (language_id);
