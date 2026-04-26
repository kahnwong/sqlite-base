-- +goose Up
ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user';

-- +goose Down
CREATE TABLE users_tmp (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

INSERT INTO users_tmp (id, name, email, created_at, updated_at)
SELECT id, name, email, created_at, updated_at
FROM users;

DROP TABLE users;
ALTER TABLE users_tmp RENAME TO users;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email);
