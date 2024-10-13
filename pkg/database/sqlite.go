package database

import (
	"database/sql"
)

type SqliteDatabaseClient struct {
	db *sql.DB
}

func (d *Database) InitSqliteDatabase() DatabaseClient {
	db, err := sql.Open("sqlite3", d.config.Path)

	if err != nil {
		d.logger.Fatal().Err(err).Msg("Could not open SQLite database")
	}

	var version string
	err = db.QueryRow("SELECT SQLITE_VERSION()").Scan(&version)

	if err != nil {
		d.logger.Fatal().Err(err).Msg("Could not query SQLite database")
	}

	d.logger.Info().
		Str("version", version).
		Str("path", d.config.Path).
		Msg("SQLite database connected")

	return &SqliteDatabaseClient{db: db}
}

func (d *Database) GetSqliteMigrations() []string {
	return []string{
		"01-blocks.sql",
		"02-notifiers.sql",
		"03-data.sql",
		"04-events.sqlite.sql",
	}
}

func (d *SqliteDatabaseClient) Exec(query string, args ...any) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

func (d *SqliteDatabaseClient) Query(query string, args ...any) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

func (d *SqliteDatabaseClient) Prepare(query string) (*sql.Stmt, error) {
	return d.db.Prepare(query)
}

func (d *SqliteDatabaseClient) QueryRow(query string, args ...any) *sql.Row {
	return d.db.QueryRow(query, args...)
}
