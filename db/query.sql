-- name: GetCache :one
SELECT data, created_at FROM cache
WHERE key = ? LIMIT 1;

-- name: InsertCache :exec
INSERT INTO cache (key, data)
VALUES (?, ?)
ON CONFLICT(key) DO UPDATE SET
    data = excluded.data,
    created_at = CURRENT_TIMESTAMP;

-- name: DeleteOld :exec
DELETE FROM cache WHERE key like ? and created_at < ?;