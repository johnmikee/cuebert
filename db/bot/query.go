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

// BotResQuery holds the configuration for the building and executing the query.
type BotResQuery struct {
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
// by any of the methods of BotResQuery below.
func (c *Config) Query() *BotResQuery {
	return &BotResQuery{
		db:  c.db,
		log: c.log,
		ctx: c.ctx,
		st:  c.st,
	}
}

// Query executes the query against the db with built query.
func (q *BotResQuery) Query() (BR, error) {
	sql, args, err := q.sql.ToSql()

	if err != nil {
		return nil, fmt.Errorf("sql generation failed %w", err)
	}

	q.log.Trace().Str("query", sql).Interface("args", args).Msg("composed sql query")

	resp := []BotResInfo{}

	rows, err := q.db.Query(
		q.ctx,
		sql, args...)
	q.log.Trace().Msg("query executed")
	if err != nil {
		return nil, fmt.Errorf("bot response query failed %w", err)
	}

	for rows.Next() {
		var br BotResInfo
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
func (q *BotResQuery) All() *BotResQuery {
	q.sql = q.st.Select("*").From(table)
	return q
}

// FirstACK queries the devices table for users who have acknowledged the first message
func (q *BotResQuery) FirstACK(serial string) *BotResQuery {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"first_ack": serial})

	return q
}

// FirstACKTime queries the devices table for users who have acknowledged the first message
func (q *BotResQuery) FirstACKTime(fm string, op compare.Compare) *BotResQuery {
	q.sql = compare.Comparison(q.st.Select("*").From(table), "first_ack_time", fm, op)

	return q
}

// FirstMessageSent queries the devices table for users who have received the first message
func (q *BotResQuery) FirstMessageSent(fm bool) *BotResQuery {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"first_message_sent": fm})

	return q
}

// DelayAt queries the devices table for users who delayed at a specific time
func (q *BotResQuery) DelayAt(d string, op compare.Compare) *BotResQuery {
	q.sql = compare.Comparison(q.st.Select("*").From(table), "delay_at", d, op)

	return q
}

// DelaySent queries the devices table for users who have received the delay message
func (q *BotResQuery) DelaySent(d string) *BotResQuery {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"delay_sent": d})

	return q
}

// SlackID queries the table for a specific slack id value
func (q *BotResQuery) SlackID(id ...string) *BotResQuery {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"slack_id": id})

	return q
}

// ManagerSlackID queries the table for a specific slack id value
func (q *BotResQuery) ManagerSlackID(id ...string) *BotResQuery {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"manager_slack_id": id})

	return q
}

// FirstMessageWaiting queries the devices table for users who have been sent the first message but not delivered
func (q *BotResQuery) FirstMessageWaiting(fm bool) *BotResQuery {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"first_message_waiting": fm})

	return q
}

// FullName queries the devices table for a specific full_name value
func (q *BotResQuery) FullName(name ...string) *BotResQuery {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"full_name": name})

	return q
}

// ManagerMessageSent queries the devices table for a specific manager_message_sent value
func (q *BotResQuery) ManagerMessageSent(sent bool) *BotResQuery {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"manager_message_sent": sent})

	return q
}

// Serial queries the devices table for a specific serial_number value
func (q *BotResQuery) Serial(serial ...string) *BotResQuery {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"serial_number": serial})

	return q
}

// TZ queries the devices table for a specific tz_offset value
func (q *BotResQuery) TZ(tz int64) *BotResQuery {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"tz_offset": tz})

	return q
}

// TZS queries the devices table multiple tz_offset values
func (q *BotResQuery) TZS(tzs ...int64) *BotResQuery {
	s := []string{}
	for _, i := range tzs {
		s = append(s, strconv.Itoa(int(i)))
	}

	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"tz_offset": s})

	return q
}

// UserEmail queries the devices table for a specific user_email value
func (q *BotResQuery) UserEmail(email ...string) *BotResQuery {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"user_email": email})

	return q
}

// Created queries the devices table for a specific created_at value
func (q *BotResQuery) Created(created string, op compare.Compare) *BotResQuery {
	q.sql = compare.Comparison(q.st.Select("*").From(table), "created_at", created, op)

	return q
}

// Created queries the devices table for a specific created_at value
func (q *BotResQuery) Updated(updated string, op compare.Compare) *BotResQuery {
	q.sql = compare.Comparison(q.st.Select("*").From(table), "updated_at", updated, op)

	return q
}
