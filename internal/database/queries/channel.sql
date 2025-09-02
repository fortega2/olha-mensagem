-- name: GetAllChannels :many
SELECT *
FROM channels
ORDER BY created_at DESC;

-- name: GetChannelByID :one
SELECT *
FROM channels
WHERE id = ?;

-- name: CreateChannel :one
INSERT INTO channels (name, description, created_by)
VALUES (?, ?, ?)
RETURNING *;

-- name: DeleteChannel :exec
DELETE FROM channels
WHERE id = ? AND created_by = ?;