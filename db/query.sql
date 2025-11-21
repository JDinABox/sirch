-- name: GetCache :one
SELECT data, expires FROM cache
WHERE key = ? LIMIT 1;

-- name: InsertCache :exec
INSERT INTO cache (key, data, expires)
VALUES (?, ?, ?)
ON CONFLICT(key) DO UPDATE SET
    data = excluded.data,
    expires = excluded.expires;

-- name: DeleteOld :exec
DELETE FROM cache WHERE expires <= ?;