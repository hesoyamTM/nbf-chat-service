CREATE TABLE IF NOT EXISTS messages (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    chat_id uuid NOT NULL,
    content text NOT NULL,
    content_type text NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_messages_user_id ON messages (user_id);
CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON messages (chat_id);

CREATE TABLE IF NOT EXISTS chats (
    id uuid DEFAULT GEN_RANDOM_UUID(),
    user_id uuid NOT NULL,
    group_id uuid,
    chat_name text NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW(),
    last_read_at timestamp NOT NULL DEFAULT NOW(),
    unread_count int NOT NULL DEFAULT 0,

    PRIMARY KEY (id, user_id),
    UNIQUE (user_id, group_id)
);

CREATE INDEX IF NOT EXISTS idx_chats_user_id ON chats (user_id);
CREATE INDEX IF NOT EXISTS idx_chats_group_id ON chats (group_id);

CREATE TABLE IF NOT EXISTS groups (
    group_id uuid PRIMARY KEY,
    chat_id uuid DEFAULT GET_RANDOM_UUID()
);
