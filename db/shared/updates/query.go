package updates

import (
	"time"

	sq "github.com/Masterminds/squirrel"
)

// Query holds the configuration for the building and executing the query.
type Query struct {
	sql    sq.SelectBuilder
	day    string
	week   int
	task   string
	config *Config
}

// By returns a new client used to interact with specific columns
// in the bot_results table.
//
// Valid options are any of the columnds in the table which can be accessed
// by any of the methods of Query below.
func (c *Config) By() *Query {
	return &Query{
		config: c,
	}
}

func (q *Query) Query() (UI, error) {
	sql, args, err := q.sql.ToSql()
	if err != nil {
		return nil, err
	}

	q.config.log.Trace().Str("query", sql).Interface("args", args).Msg("composed sql query")

	resp := []Info{}

	rows, err := q.config.db.Query(
		q.config.ctx,
		sql, args...)
	q.config.log.Trace().Msg("query executed")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var ui Info
		err = rows.Scan(
			&ui.Week,
			&ui.ItemName,
			&ui.Monday,
			&ui.Tuesday,
			&ui.Wednesday,
			&ui.Thursday,
			&ui.Friday,
			&ui.UpdatedAt,
		)
		if err != nil {
			q.config.log.Info().AnErr("err", err).Msg("failed to scan row")
			return nil, err
		}

		resp = append(resp, ui)
	}

	return resp, nil
}

type Days string

const (
	Monday    Days = "monday"
	Tuesday   Days = "tuesday"
	Wednesday Days = "wednesday"
	Thursday  Days = "thursday"
	Friday    Days = "friday"
	Today     Days = "today"
)

func (q *Query) Day(day Days) *Query {
	if day != Today {
		q.day = string(day)
		return q
	}

	// get the day of the week and switch to set it
	// to the correct day
	today := time.Now().Weekday()
	switch today {
	case time.Monday:
		day = Monday
	case time.Tuesday:
		day = Tuesday
	case time.Wednesday:
		day = Wednesday
	case time.Thursday:
		day = Thursday
	case time.Friday:
		day = Friday
	}
	q.day = string(day)

	return q
}

func (q *Query) Week(week int) *Query {
	q.week = week
	return q
}

func (q *Query) Task(task string) *Query {
	q.task = task
	return q
}

func (q *Query) Build() *Query {
	// var base sq.SelectBuilder
	// if q.day == "" {
	// 	base = q.config.st.Select(columns...).From(q.config.table)
	// } else {
	// 	base = q.config.st.Select(q.day).From(q.config.table)
	// }
	base := q.config.st.Select(columns...).From(q.config.table)

	if q.week != 0 && q.task != "" {
		q.sql = base.Where(sq.Eq{"week": q.week, "item_name": q.task})
		return q
	}

	if q.week != 0 {
		q.sql = base.Where(sq.Eq{"week": q.week})
		return q
	}

	if q.task != "" {
		q.sql = base.Where(sq.Eq{"item_name": q.task})
		return q
	}

	q.sql = base

	return q
}
