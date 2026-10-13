CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    role TEXT NOT NULL DEFAULT 'user'
);

CREATE UNIQUE INDEX idx_users_email ON users (email);
