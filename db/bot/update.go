package bot

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/db/parser"
)

// Update initializes a new BotResUpdate struct.
//
// The methods of BotResUpdate are used to modify which values
// will be updated.
//
// Those are parsed and the column we are using as the condition to match is passed as the index.
func (c *Config) Update() *BotResUpdate {
	return &BotResUpdate{
		db:  c.db,
		ctx: context.Background(),
		log: c.log,
	}
}

// Update sends the statement to update the device after it has been composed.
// returns the connection which should be closed after checking the error.
func (u *BotResUpdate) Send() (*pgxpool.Conn, error) {
	u.log.Trace().Str("query", u.query).Interface("args", u.args).Msg("composed sql query")
	_, err := u.db.Exec(
		u.ctx,
		u.query, u.args...)

	if err != nil {
		return nil, err
	}

	u.db.Release()
	u.log.Trace().Str("table", table).Msg("input was successfully submitted")

	return u.db, nil
}

// ParseInput will take the input provided by the user via the methods
// of BotResUpdate and compose a statement.
//
// As arguments are added they are sorted alphabetically. This is by no
// means a foolproof way of sorting the data but given the small subset of
// columns in our table this will work to compose the arguments before sending
// it to postgres to be executed.
func (u *BotResUpdate) Parse(index, val string) *BotResUpdate {
	check := []parser.CheckInfo{
		{
			Fn: parser.Prim{
				S: u.bresp.SlackID,
			},
			Key:     "slack_id",
			Trimmed: "SlackID",
		},
		{
			Fn: parser.Prim{
				S: u.bresp.UserEmail,
			},
			Key:     "user_email",
			Trimmed: "UserEmail",
		},
		{
			Fn: parser.Prim{
				S: u.bresp.ManagerSlackID,
			},
			Key:     "manager_slack_id",
			Trimmed: "ManagerSlackID",
		},
		{
			Fn: parser.Prim{
				S: u.bresp.SerialNumber,
			},
			Key:     "serial_number",
			Trimmed: "SerialNumber",
		},
		{
			Fn: parser.Prim{
				B: u.bresp.FirstACK,
			},
			Key:     "first_ack",
			Trimmed: "FirstACK",
		},
		{
			Fn: parser.Prim{
				T: u.bresp.FirstACKTime,
			},
			Key:     "first_ack_time",
			Trimmed: "FirstACKTime",
		},
		{
			Fn: parser.Prim{
				B: u.bresp.FirstMessageSent,
			},
			Key:     "first_message_sent",
			Trimmed: "FirstMessageSent",
		},
		{
			Fn: parser.Prim{
				T: u.bresp.FirstMessageSentAt,
			},
			Key:     "first_message_sent_at",
			Trimmed: "FirstMessageSentAt",
		},
		{
			Fn: parser.Prim{
				B: u.bresp.FirstMessageWaiting,
			},
			Key:     "first_message_waiting",
			Trimmed: "FirstMessageWaiting",
		},
		{
			Fn: parser.Prim{
				S: u.bresp.FullName,
			},
			Key:     "full_name",
			Trimmed: "FullName",
		},
		{
			Fn: parser.Prim{
				B: u.bresp.ManagerMessageSent,
			},
			Key:     "manager_message_sent",
			Trimmed: "ManagerMessageSent",
		},
		{
			Fn: parser.Prim{
				T: u.bresp.ManagerMessageSentAt,
			},
			Key:     "manager_message_sent_at",
			Trimmed: "ManagerMessageSentAt",
		},
		{
			Fn: parser.Prim{
				T: u.bresp.DelayAt,
			},
			Key:     "delay_at",
			Trimmed: "DelayAt",
		},
		{
			Fn: parser.Prim{
				S: u.bresp.DelayDate,
			},
			Key:     "delay_date",
			Trimmed: "DelayDate",
		},
		{
			Fn: parser.Prim{
				S: u.bresp.DelayTime,
			},
			Key:     "delay_time",
			Trimmed: "DelayTime",
		},
		{
			Fn: parser.Prim{
				B: u.bresp.DelaySent,
			},
			Key:     "delay_sent",
			Trimmed: "DelaySent",
		},
	}

	query, args, err := parser.ParseInput(&parser.Parser{
		Index:  index,
		Table:  table,
		Val:    val,
		Check:  check,
		Method: parser.Update,
		Into:   BotResInfo{},
	})
	if err != nil {
		u.log.Err(err).Msg("parse input")
		return nil
	}

	u.args = args
	u.query = query

	return u
}
