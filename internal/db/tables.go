package db

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/johnmikee/cuebert/db"
	"github.com/johnmikee/cuebert/pkg/helpers"
)

func (c *Config) Delete(tables []string) error {
	for _, t := range tables {
		if !helpers.Contains(db.CueTables, t) {
			return fmt.Errorf("%s is not a table", t)
		}

		query, args, err := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
			Delete(t).
			ToSql()

		c.log.Trace().
			Str("query", query).
			Str("args", fmt.Sprintf("%v", args)).
			Str("table", t).
			Msg("delete query")

		if err != nil {
			return fmt.Errorf("building delete query for %s failed: %s", t, err)
		}

		_, err = c.db.Exec(context.Background(), query, args...)
		if err != nil {
			return fmt.Errorf("deleting table %s failed: %s", t, err)
		}
	}

	return nil
}
