-- name: CreateMessage :exec
INSERT INTO messages (channel_id, user_id, user_color, content)
VALUES (?, ?, ?, ?);

-- name: GetHistoryMessagesByChannel :many
SELECT
    m.id,
    m.channel_id,
    m.user_id,
    m.user_color,
    u.username AS user_username,
    m.content,
    m.created_at
FROM
    messages AS m
INNER JOIN
    users AS u ON u.id = m.user_id
WHERE
    m.channel_id = ?
ORDER BY
    m.created_at ASC
LIMIT
    ?;
