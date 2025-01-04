package database

import (
	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"
)

type StubDatabaseClient struct {
	MigrateError error
	ExecError    error
	Client       *sql.DB
	Mock         sqlmock.Sqlmock
}

func NewStubDatabaseClient() *StubDatabaseClient {
	db, mock, _ := sqlmock.New()
	return &StubDatabaseClient{Client: db, Mock: mock}
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
	return d.Client.Query(query, args...)
}

func (d *StubDatabaseClient) QueryRow(query string, args ...any) *sql.Row {
	return d.Client.QueryRow(query, args...)
}

func (d *StubDatabaseClient) Migrate() error {
	return d.MigrateError
}
