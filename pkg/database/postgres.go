package database

import (
	"database/sql"
	postgresMigrations "main/migrations/postgres"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

type PostgresDatabaseClient struct {
	db     *sql.DB
	logger *DatabaseLogger
}

func (d *Database) InitPostgresDatabase() DatabaseClient {
	db, err := sql.Open("postgres", d.config.Path)

	if err != nil {
		d.logger.Panic().Err(err).Msg("Could not open PostgreSQL database")
	}

	var version string
	err = db.QueryRow("SELECT version()").Scan(&version)

	if err != nil {
		d.logger.Panic().Err(err).Msg("Could not query PostgreSQL database")
	}

	d.logger.Info().
		Str("version", version).
		Str("path", d.config.Path).
		Msg("PostgreSQL database connected")

	return &PostgresDatabaseClient{db: db, logger: NewDatabaseLogger(d.logger)}
}

func (d *PostgresDatabaseClient) Exec(query string, args ...any) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

func (d *PostgresDatabaseClient) Query(query string, args ...any) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

func (d *PostgresDatabaseClient) QueryRow(query string, args ...any) *sql.Row {
	return d.db.QueryRow(query, args...)
}

func (d *PostgresDatabaseClient) Migrate() error {
	goose.SetBaseFS(postgresMigrations.EmbedFS)
	goose.SetLogger(d.logger)

	_ = goose.SetDialect("postgres")

	return goose.Up(d.db, ".")
}
