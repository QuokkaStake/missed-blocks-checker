package database

import "database/sql"

type StubDatabaseClient struct {
	MigrateError error
	ExecError    error
}

type StubSQLResult struct{}

func (s *StubSQLResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (s *StubSQLResult) RowsAffected() (int64, error) {
	return 0, nil
}

func (d *StubDatabaseClient) Exec(query string, args ...any) (sql.Result, error) {
	return &StubSQLResult{}, d.ExecError
}

func (d *StubDatabaseClient) Query(query string, args ...any) (*sql.Rows, error) {
	return nil, d.MigrateError
}

func (d *StubDatabaseClient) QueryRow(query string, args ...any) *sql.Row {
	return nil
}

func (d *StubDatabaseClient) Migrate() error {
	return d.MigrateError
}
