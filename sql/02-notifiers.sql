CREATE TABLE IF NOT EXISTS notifiers (
    chain TEXT NOT NULL,
    reporter TEXT NOT NULL,
    operator_address TEXT NOT NULL,
    user_name TEXT NOT NULL,
    user_id TEXT NOT NULL,
    PRIMARY KEY (chain, reporter, operator_address, user_id)
);