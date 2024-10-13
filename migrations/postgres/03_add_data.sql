-- +goose Up
CREATE TABLE IF NOT EXISTS data (
    chain TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    PRIMARY KEY (chain, key)
);

-- +goose Down
DROP TABLE data;
