-- name: CreateUser :exec
INSERT INTO users(name, email, role, created_at, updated_at)
VALUES (sqlc.arg(name), sqlc.arg(email), sqlc.arg(role), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- name: CountUsers :one
SELECT COUNT(1) FROM users;
