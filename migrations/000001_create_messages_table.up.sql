CREATE TABLE IF NOT EXISTS messages (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    group_id uuid NOT NULL,
    message text NOT NULL,
    created_at timestamp NOT NULL DEFAULT NOW()
);
