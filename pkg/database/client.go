package database

import (
	"database/sql"
	"strings"

	"github.com/rs/zerolog"
)

type DatabaseLogger struct {
	Logger zerolog.Logger
}

func NewDatabaseLogger(logger zerolog.Logger) *DatabaseLogger {
	return &DatabaseLogger{
		Logger: logger.With().Str("component", "database_migrations").Logger(),
	}
}

func (l *DatabaseLogger) Printf(format string, v ...interface{}) {
	l.Logger.Info().Msgf(strings.TrimSpace(format), v...)
}

func (l *DatabaseLogger) Fatalf(format string, v ...interface{}) {
	l.Logger.Panic().Msgf(strings.TrimSpace(format), v...)
}

type DatabaseClient interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	Migrate() error
}
