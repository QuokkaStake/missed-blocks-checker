CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    chain TEXT NOT NULL,
    height BIGINT NOT NULL,
    event TEXT NOT NULL,
    validator TEXT NOT NULL,
    payload TEXT NOT NULL,
    time TIMESTAMP NOT NULL
);