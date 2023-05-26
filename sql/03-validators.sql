CREATE TABLE IF NOT EXISTS validators (
      chain TEXT NOT NULL,
      height BIGINT NOT NULL,
      validators TEXT NOT NULL,
      PRIMARY KEY (chain, height, validators)
);