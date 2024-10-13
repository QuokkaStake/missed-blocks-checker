package database

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type PostgresDatabaseClient struct {
	db *sql.DB
}

func (d *Database) InitPostgresDatabase() DatabaseClient {
	db, err := sql.Open("postgres", d.config.Path)

	if err != nil {
		d.logger.Fatal().Err(err).Msg("Could not open PostgreSQL database")
	}

	var version string
	err = db.QueryRow("SELECT version()").Scan(&version)

	if err != nil {
		d.logger.Fatal().Err(err).Msg("Could not query PostgreSQL database")
	}

	d.logger.Info().
		Str("version", version).
		Str("path", d.config.Path).
		Msg("PostgreSQL database connected")

	return &PostgresDatabaseClient{db: db}
}

func (d *Database) GetPostgresMigrations() []string {
	return []string{
		"01-blocks.sql",
		"02-notifiers.sql",
		"03-data.sql",
		"04-events.postgres.sql",
	}
}

func (d *PostgresDatabaseClient) Exec(query string, args ...any) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

func (d *PostgresDatabaseClient) Query(query string, args ...any) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

func (d *PostgresDatabaseClient) Prepare(query string) (*sql.Stmt, error) {
	return d.db.Prepare(query)
}

func (d *PostgresDatabaseClient) QueryRow(query string, args ...any) *sql.Row {
	return d.db.QueryRow(query, args...)
}
