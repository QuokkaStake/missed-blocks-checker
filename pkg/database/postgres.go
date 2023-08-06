package database

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func (d *Database) InitPostgresDatabase() *sql.DB {
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

	return db
}

func (d *Database) GetPostgresMigrations() []string {
	return []string{
		"01-blocks.sql",
		"02-notifiers.sql",
		"03-data.sql",
		"04-events.postgres.sql",
	}
}
