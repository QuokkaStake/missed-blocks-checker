CREATE TABLE IF NOT EXISTS notifiers (
    chain TEXT NOT NULL,
    reporter TEXT NOT NULL,
    operator_address TEXT NOT NULL,
    notifier TEXT NOT NULL,
    PRIMARY KEY (chain, reporter, operator_address, notifier)
);