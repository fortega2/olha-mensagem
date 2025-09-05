-- name: GetAllChannels :many
SELECT
    c.id,
    c.name,
    c.description,
    c.created_by,
    u.username AS created_by_username,
    c.created_at
FROM
    channels AS c
INNER JOIN
    users AS u ON u.id = c.created_by
ORDER BY
    c.id DESC;

-- name: GetChannelByID :one
SELECT
    c.id,
    c.name,
    c.description,
    c.created_by,
    u.username AS created_by_username,
    c.created_at
FROM
    channels AS c
INNER JOIN
    users AS u ON u.id = c.created_by
WHERE
    c.id = ?;

-- name: CreateChannel :one
INSERT INTO channels (name, description, created_by)
VALUES (?, ?, ?)
RETURNING id;

-- name: DeleteChannel :exec
DELETE FROM channels
WHERE id = ? AND created_by = ?;