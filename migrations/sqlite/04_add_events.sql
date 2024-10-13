-- +goose Up
CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    chain TEXT NOT NULL,
    height BIGINT NOT NULL,
    event TEXT NOT NULL,
    validator TEXT NOT NULL,
    payload TEXT NOT NULL,
    time TEXT NOT NULL
);

-- +goose Down
DROP TABLE events;
