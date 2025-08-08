package overdue

import sq "github.com/Masterminds/squirrel"

type Query struct {
	config *Config
	sql    sq.SelectBuilder
}

func (c *Config) By() *Query {
	return &Query{
		config: c,
	}
}

func (q *Query) Query() (UI, error) {
	sql, args, err := q.config.st.Select(columns...).From(q.config.table).ToSql()
	if err != nil {
		return nil, err
	}

	resp := []Info{}

	rows, err := q.config.db.Query(
		q.config.ctx,
		sql, args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var ui Info
		err = rows.Scan(
			&ui.UserID,
			&ui.UserServiceID,
			&ui.UserEmail,
			&ui.Task,
		)
		if err != nil {
			return nil, err
		}

		resp = append(resp, ui)
	}

	return resp, nil
}

// All returns all users and values in the table.
func (q *Query) All() *Query {
	q.sql = q.config.st.Select(columns...).From(q.config.table)

	return q
}

func (q *Query) sqeq(column string, value interface{}) *Query {
	q.sql = q.config.st.Select(columns...).From(q.config.table).Where(sq.Eq{column: value})

	return q
}

func (q *Query) Opts(column Column, val any) *Query {
	return q.sqeq(string(column), val)
}
