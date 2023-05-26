CREATE TABLE IF NOT EXISTS validators (
      chain TEXT NOT NULL,
      height BIGINT NOT NULL,
      validator_address TEXT NOT NULL,
      PRIMARY KEY (chain, height, validator_address)
);