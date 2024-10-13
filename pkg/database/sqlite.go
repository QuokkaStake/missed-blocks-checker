package database

import (
	"database/sql"
	sqliteMigrations "main/migrations/sqlite"

	"github.com/pressly/goose/v3"
)

type SqliteDatabaseClient struct {
	db     *sql.DB
	logger *DatabaseLogger
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

	return &SqliteDatabaseClient{db: db, logger: NewDatabaseLogger(d.logger)}
}

func (d *SqliteDatabaseClient) Exec(query string, args ...any) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

func (d *SqliteDatabaseClient) Query(query string, args ...any) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

func (d *SqliteDatabaseClient) QueryRow(query string, args ...any) *sql.Row {
	return d.db.QueryRow(query, args...)
}

func (d *SqliteDatabaseClient) Migrate() error {
	goose.SetBaseFS(sqliteMigrations.EmbedFS)
	goose.SetLogger(d.logger)

	_ = goose.SetDialect("sqlite")

	return goose.Up(d.db, ".")
}
