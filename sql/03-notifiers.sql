CREATE TABLE IF NOT EXISTS notifiers (
      reporter TEXT NOT NULL,
      operator_address TEXT NOT NULL,
      notifier TEXT NOT NULL,
      PRIMARY KEY (reporter, operator_address, notifier)
);