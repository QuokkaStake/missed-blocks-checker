package database

import (
	"database/sql"
)

func (d *Database) InitSqliteDatabase() *sql.DB {
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

	return db
}

func (d *Database) GetSqliteMigrations() []string {
	return []string{
		"01-blocks.sql",
		"02-notifiers.sql",
		"03-data.sql",
		"04-events.sqlite.sql",
	}
}
