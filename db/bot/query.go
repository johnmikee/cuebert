package bot

import (
	"context"
	"fmt"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/db/compare"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// Query holds the configuration for the building and executing the query.
type Query struct {
	ctx context.Context
	sql sq.SelectBuilder
	st  sq.StatementBuilderType

	db  *pgxpool.Conn
	log logger.Logger
}

// By returns a new client used to interact with specific columns
// in the bot_results table.
//
// Valid options are any of the columnds in the table which can be accessed
// by any of the methods of Query below.
func (c *Config) Query() *Query {
	return &Query{
		db:  c.db,
		log: c.log,
		ctx: c.ctx,
		st:  c.st,
	}
}

// Query executes the query against the db with built query.
func (q *Query) Query() (BR, error) {
	sql, args, err := q.sql.ToSql()

	if err != nil {
		return nil, fmt.Errorf("sql generation failed %w", err)
	}

	q.log.Trace().Str("query", sql).Interface("args", args).Msg("composed sql query")

	resp := []Info{}

	rows, err := q.db.Query(
		q.ctx,
		sql, args...)
	q.log.Trace().Msg("query executed")
	if err != nil {
		return nil, fmt.Errorf("bot response query failed %w", err)
	}

	for rows.Next() {
		var br Info
		err = rows.Scan(
			&br.SlackID,
			&br.UserEmail,
			&br.ManagerSlackID,
			&br.FirstACK,
			&br.FirstACKTime,
			&br.FirstMessageSent,
			&br.FirstMessageSentAt,
			&br.FirstMessageWaiting,
			&br.ManagerMessageSent,
			&br.ManagerMessageSentAt,
			&br.FullName,
			&br.DelayAt,
			&br.DelayDate,
			&br.DelayTime,
			&br.DelaySent,
			&br.ReminderInterval,
			&br.ReminderWaiting,
			&br.SerialNumber,
			&br.TZOffset,
			&br.CreatedAt,
			&br.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("bot row query failed %w", err)
		}
		resp = append(resp, br)
	}

	q.db.Release()

	return resp, nil
}

// All returns all results in the table
func (q *Query) All() *Query {
	q.sql = q.st.Select("*").From(table)
	return q
}

// FirstACK queries the devices table for users who have acknowledged the first message
func (q *Query) FirstACK(serial string) *Query {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"first_ack": serial})

	return q
}

// FirstACKTime queries the devices table for users who have acknowledged the first message
func (q *Query) FirstACKTime(fm string, op compare.Compare) *Query {
	q.sql = compare.Comparison(q.st.Select("*").From(table), "first_ack_time", fm, op)

	return q
}

// FirstMessageSent queries the devices table for users who have received the first message
func (q *Query) FirstMessageSent(fm bool) *Query {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"first_message_sent": fm})

	return q
}

// DelayAt queries the devices table for users who delayed at a specific time
func (q *Query) DelayAt(d string, op compare.Compare) *Query {
	q.sql = compare.Comparison(q.st.Select("*").From(table), "delay_at", d, op)

	return q
}

// DelaySent queries the devices table for users who have received the delay message
func (q *Query) DelaySent(d string) *Query {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"delay_sent": d})

	return q
}

// SlackID queries the table for a specific slack id value
func (q *Query) SlackID(id ...string) *Query {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"slack_id": id})

	return q
}

// ManagerSlackID queries the table for a specific slack id value
func (q *Query) ManagerSlackID(id ...string) *Query {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"manager_slack_id": id})

	return q
}

// FirstMessageWaiting queries the devices table for users who have been sent the first message but not delivered
func (q *Query) FirstMessageWaiting(fm bool) *Query {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"first_message_waiting": fm})

	return q
}

// FullName queries the devices table for a specific full_name value
func (q *Query) FullName(name ...string) *Query {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"full_name": name})

	return q
}

// ManagerMessageSent queries the devices table for a specific manager_message_sent value
func (q *Query) ManagerMessageSent(sent bool) *Query {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"manager_message_sent": sent})

	return q
}

// Serial queries the devices table for a specific serial_number value
func (q *Query) Serial(serial ...string) *Query {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"serial_number": serial})

	return q
}

// ReminderInterval queries the devices table for a specific reminder_interval value
func (q *Query) ReminderInterval(interval ...int) *Query {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"reminder_interval": interval})

	return q
}

// ReminderWaiting queries the devices table for a devices with reminders already sent
func (q *Query) ReminderWaiting(waiting bool) *Query {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"reminder_waiting": waiting})

	return q
}

// TZ queries the devices table for a specific tz_offset value
func (q *Query) TZ(tz int64) *Query {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"tz_offset": tz})

	return q
}

// TZS queries the devices table multiple tz_offset values
func (q *Query) TZS(tzs ...int64) *Query {
	s := []string{}
	for _, i := range tzs {
		s = append(s, strconv.Itoa(int(i)))
	}

	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"tz_offset": s})

	return q
}

// UserEmail queries the devices table for a specific user_email value
func (q *Query) UserEmail(email ...string) *Query {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"user_email": email})

	return q
}

// Created queries the devices table for a specific created_at value
func (q *Query) Created(created string, op compare.Compare) *Query {
	q.sql = compare.Comparison(q.st.Select("*").From(table), "created_at", created, op)

	return q
}

// Created queries the devices table for a specific created_at value
func (q *Query) Updated(updated string, op compare.Compare) *Query {
	q.sql = compare.Comparison(q.st.Select("*").From(table), "updated_at", updated, op)

	return q
}
