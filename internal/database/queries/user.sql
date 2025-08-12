-- name: CreateUser :one
INSERT INTO users (username, password)
VALUES (?, ?)
RETURNING *;

-- name: GetUserByUsernameAndPassword :one
SELECT *
FROM users
WHERE username = ?
AND password = ?;