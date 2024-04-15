package tracking

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// Add contains the information for adding a new item to the tracking table.
type Add struct {
	ui     Info
	config *Config
}

// Add returns a new Add struct.
func (c *Config) Add() *Add {
	return &Add{
		config: c,
		ui:     Info{},
	}
}

// AddAll adds all the items to the database.
func (u *Add) All(b TI) (int64, error) {
	for i := range b {
		sql, args, err := u.config.st.Insert(u.config.table).
			Columns(columns...).
			Values(
				b[i].SerialNumber,
				b[i].UserServiceID,
				b[i].TaskComplete,
				b[i].TaskName,
				b[i].UserSlackID,
				b[i].UserEmail,
				b[i].UserFullName,
				b[i].FirstACK,
				b[i].FirstACKTime,
				b[i].FirstMessageSent,
				b[i].FirstMessageSentAt,
				b[i].FirstMessageWaiting,
				b[i].UserReminderAt,
				b[i].UserReminderDate,
				b[i].UserReminderTime,
				b[i].ReminderSent,
				b[i].ReminderInterval,
				b[i].ReminderWaiting,
				b[i].TZOffset,
				b[i].UpdatedAt,
				b[i].FinalMessageSent,
				b[i].FinalMessageSentAt,
			).Suffix("ON CONFLICT DO NOTHING").ToSql()
		if err != nil {
			u.config.log.Err(err).Msg("sql generation failed")
			continue
		}

		_, err = u.config.db.Exec(context.Background(), sql, args...)
		if err != nil {
			u.config.log.Err(err).Msg("exec failed")
			continue
		}
	}

	return 0, nil
}

func (u *Add) Batch(batch *pgx.Batch) {
	br := u.config.db.SendBatch(context.Background(), batch)
	ct, err := br.Exec()
	if err != nil {
		u.config.log.Err(err).Msg("exec batch")
		return
	}

	u.config.log.Debug().Int64("count", ct.RowsAffected()).Msg("rows affected")
}

func (u *Add) Partial(b TI, cols []Column) error {
	rows := [][]interface{}{}
	columns := colStrArr(cols)

	for i := range b {
		row := []interface{}{}
		for _, v := range cols {
			switch v {
			case SerialNumber:
				row = append(row, b[i].SerialNumber)
			case UserServiceID:
				row = append(row, b[i].UserServiceID)
			case TaskComplete:
				row = append(row, b[i].TaskComplete)
			case TaskName:
				row = append(row, b[i].TaskName)
			case UserSlackID:
				row = append(row, b[i].UserSlackID)
			case UserEmail:
				row = append(row, b[i].UserEmail)
			case UserFullName:
				row = append(row, b[i].UserFullName)
			case FirstACK:
				row = append(row, b[i].FirstACK)
			case FirstACKTime:
				row = append(row, b[i].FirstACKTime)
			case FirstMessageSent:
				row = append(row, b[i].FirstMessageSent)
			case FirstMessageSentAt:
				row = append(row, b[i].FirstMessageSentAt)
			case FirstMessageWaiting:
				row = append(row, b[i].FirstMessageWaiting)
			case UserReminderAt:
				row = append(row, b[i].UserReminderAt)
			case UserReminderDate:
				row = append(row, b[i].UserReminderDate)
			case UserReminderTime:
				row = append(row, b[i].UserReminderTime)
			case ReminderSent:
				row = append(row, b[i].ReminderSent)
			case ReminderInterval:
				row = append(row, b[i].ReminderInterval)
			case ReminderWaiting:
				row = append(row, b[i].ReminderWaiting)
			case TZOffset:
				row = append(row, b[i].TZOffset)
			case UpdatedAt:
				row = append(row, b[i].UpdatedAt)
			case FinalMessageSent:
				row = append(row, b[i].FinalMessageSent)
			case FinalMessageSentAt:
				row = append(row, b[i].FinalMessageSentAt)
			}
		}
		rows = append(rows, row)
	}

	return u.batchall(columns, rows)
}

func colStrArr(col []Column) []string {
	var columns []string
	for _, v := range col {
		columns = append(columns, string(v))
	}
	return columns
}

func (u *Add) batchall(columns []string, rows [][]interface{}) error {
	batch := &pgx.Batch{}
	for i := range rows {
		sql, args, err := u.config.st.Insert(u.config.table).Columns(columns...).Values(rows[i]...).Suffix("ON CONFLICT DO NOTHING").ToSql()
		if err != nil {
			u.config.log.Err(err).Msg("sql generation failed")
			continue
		}
		batch.Queue(sql, args...)
	}

	br := u.config.db.SendBatch(context.Background(), batch)
	ct, err := br.Exec()
	if err != nil {
		u.config.log.Err(err).Msg("exec batch")
		return err
	}

	u.config.log.Debug().Int64("count", ct.RowsAffected()).Msg("rows affected")
	return nil
}

// Execute takes the prepared statement and executes it.
func (u *Add) Execute() error {
	query, args, err := u.config.st.Insert(u.config.table).
		Columns(columns...).
		Values(
			u.ui.SerialNumber,
			u.ui.UserServiceID,
			u.ui.TaskComplete,
			u.ui.TaskName,
			u.ui.UserSlackID,
			u.ui.UserEmail,
			u.ui.UserFullName,
			u.ui.FirstACK,
			u.ui.FirstACKTime,
			u.ui.FirstMessageSent,
			u.ui.FirstMessageSentAt,
			u.ui.FirstMessageWaiting,
			u.ui.UserReminderAt,
			u.ui.UserReminderDate,
			u.ui.UserReminderTime,
			u.ui.ReminderSent,
			u.ui.ReminderInterval,
			u.ui.ReminderWaiting,
			u.ui.TZOffset,
			u.ui.UpdatedAt,
		).
		ToSql()
	if err != nil {
		return err
	}

	u.config.log.Trace().Str("query", query).Interface("args", args).Msg("query")

	_, err = u.config.db.Exec(
		u.config.ctx,
		query, args...)
	if err != nil {
		return err
	}

	u.config.db.Release()
	u.config.log.Info().Msg("result was successfully added")

	return nil
}

// WithOptions sets the options for the Add struct.
func (a *Add) WithOptions(i Info) *Add {
	a.ui = i

	return a
}
