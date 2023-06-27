package exclusions

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/db/parser"
)

// Update initializes a new ExclusionUpdate struct.
//
// The methods of ExclusionUpdate are used to modify which values
// will be updated.
//
// Those are parsed and the column we are using as the condition to match is passed as the index.
func (c *Config) Update() *ExclusionUpdate {
	return &ExclusionUpdate{
		db:  c.db,
		ctx: context.Background(),
		log: c.log,
		st:  c.st,
	}
}

// Update sends the statement to update the device after it has been composed.
//
// returns the connection which should be closed after checking the error.
func (u *ExclusionUpdate) Send() (*pgxpool.Conn, error) {
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
func (u *ExclusionUpdate) Parse(index, val string) *ExclusionUpdate {
	check := []parser.CheckInfo{
		{
			Fn: parser.Prim{
				B: u.e.Approved,
			},
			Key:     "approved",
			Trimmed: "Approved",
		},
		{
			Fn: parser.Prim{
				S: u.e.SerialNumber,
			},
			Key:     "serial_number",
			Trimmed: "SerialNumber",
		},
		{
			Fn: parser.Prim{
				S: u.e.Reason,
			},
			Key:     "reason",
			Trimmed: "Reason",
		},
		{
			Fn: parser.Prim{
				S: u.e.UserEmail,
			},
			Key:     "user_email",
			Trimmed: "UserEmail",
		},
		{
			Fn: parser.Prim{
				T: u.e.Until,
			},
			Key:     "until",
			Trimmed: "Until",
		},
	}

	query, args, err := parser.ParseInput(&parser.Parser{
		Index:  index,
		Table:  table,
		Val:    val,
		Check:  check,
		Method: parser.Update,
		Into:   ExclusionInfo{},
	})
	if err != nil {
		u.log.Err(err).Msg("parse input")
		return nil
	}

	u.args = args
	u.query = query

	return u
}
