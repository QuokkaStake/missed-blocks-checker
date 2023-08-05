CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    chain TEXT NOT NULL,
    event TEXT NOT NULL,
    validator TEXT NOT NULL,
    payload TEXT NOT NULL
);