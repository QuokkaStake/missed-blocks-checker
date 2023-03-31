CREATE TABLE IF NOT EXISTS signatures (
    chain TEXT NOT NULL,
    height BIGINT NOT NULL,
    validator_address TEXT NOT NULL,
    signature INT NOT NULL,
    PRIMARY KEY (chain, height, validator_address)
);