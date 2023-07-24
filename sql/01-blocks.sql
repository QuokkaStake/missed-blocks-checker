CREATE TABLE IF NOT EXISTS blocks (
    chain TEXT not null,
    height BIGINT NOT NULL,
    time BIGINT NOT NULL,
    proposer TEXT NOT NULL,
    signatures TEXT NOT NULL,
    validators TEXT NOT NULL,
    PRIMARY KEY (chain, height)
);