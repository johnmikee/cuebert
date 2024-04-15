package migrations

import (
	"embed"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" //nolint: staticcheck
	_ "github.com/golang-migrate/migrate/v4/source/file"       //nolint: staticcheck
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Make migrations referencing a local file system
func Make(connStr string, migrationFiles string) error {
	m, err := migrate.New(
		"file://"+migrationFiles,
		connStr,
	)
	if err != nil {
		return errors.Wrap(err, "failed to make migrations")
	}

	return up(m)
}

//go:embed sql/*.sql
var fs embed.FS

// MakeEmbed uses the embedded file system to make migrations of the neighboring sql directory
func MakeEmbed(connStr string) error {
	entries, err := fs.ReadDir("sql")
	if err != nil {
		return err
	}
	for _, e := range entries {
		log.Info().Str("file", e.Name()).Msg("found migration file")
	}
	d, err := iofs.New(fs, "sql")
	if err != nil {
		return errors.Wrap(err, "failed to embed migrations")
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, connStr)
	if err != nil {
		return errors.Wrap(err, "failed to make migrations")
	}

	return up(m)
}

// up is a kids movie that gives you the sad feels
func up(m *migrate.Migrate) error {
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return errors.Wrap(err, "failed to migrate")
	}

	return nil
}
