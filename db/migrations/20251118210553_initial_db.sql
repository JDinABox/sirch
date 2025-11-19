-- migrate:up
CREATE TABLE cache (
    key TEXT PRIMARY KEY,
    data BLOB NOT NULL,
    expires DATETIME NOT NULL
);

-- migrate:down
DROP TABLE cache;