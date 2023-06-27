package users

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/db/parser"
)

// UpdateDevice initializes a new UserUpdate struct.
//
// The methods of UserUpdate are used to modify which values
// will be updated. those are parsed and the
// column we are using as the condition to match is passed as the index.
func (c *Config) UpdateDevice() *UserUpdate {
	return &UserUpdate{
		db:  c.db,
		ctx: c.ctx,
		log: c.log,
		st:  c.st,
	}
}

// Update sends the statement to update the device after it has been composed.
//
// returns the connection which should be closed after checking the error.
func (u *UserUpdate) Send() (*pgxpool.Conn, error) {
	_, err := u.db.Exec(
		u.ctx,
		u.query, u.args...)

	if err != nil {
		return nil, err
	}

	u.db.Release()
	u.log.Info().Msg("input was successfully submitted")

	return u.db, nil
}

// Parse will take the input provided by the user via the methods
// of UserUpdate and compose a statement.
//
// As arguments are added they are sorted alphabetically. This is by no
// means a foolproof way of sorting the data but given the small subset of
// columns in our table this will work to compose the arguments before sending
// it to postgres to be executed.
func (u *UserUpdate) Parse(index, val string) *UserUpdate {
	check := []parser.CheckInfo{
		{
			Fn: parser.Prim{
				S: u.user.MDMID,
			},
			Key:     "id",
			Trimmed: "ID",
		},
		{
			Fn: parser.Prim{
				S: u.user.UserEmail,
			},
			Key:     "user_email",
			Trimmed: "UserEmail",
		},
		{
			Fn: parser.Prim{
				S: u.user.UserLongName,
			},
			Key:     "user_long_name",
			Trimmed: "UserLongName",
		},
		{
			Fn: parser.Prim{
				S: u.user.UserSlackID,
			},
			Key:     "user_slack_id",
			Trimmed: "UserSlackID",
		},
		{
			Fn: parser.Prim{
				I64: u.user.TZOffset,
			},
			Key:     "tz_offset",
			Trimmed: "TZOffset",
		},
	}

	query, args, err := parser.ParseInput(&parser.Parser{
		Index:  index,
		Table:  table,
		Val:    val,
		Check:  check,
		Method: parser.Update,
		Into:   UserInfo{},
	})

	if err != nil {
		u.log.Err(err).Msg("parse input")
		return nil
	}

	u.query = query
	u.args = args

	return u
}
