CREATE TABLE IF NOT EXISTS signatures (
    height BIGINT NOT NULL,
    validator_address TEXT NOT NULL,
    signed bool NOT NULL,
    PRIMARY KEY (height, validator_address)
);