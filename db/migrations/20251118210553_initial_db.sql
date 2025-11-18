-- migrate:up
CREATE TABLE cache (
    key TEXT PRIMARY KEY,
    data BLOB NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- migrate:down
DROP TABLE cache;