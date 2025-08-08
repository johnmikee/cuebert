package shared

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/johnmikee/cuebert/cuebert/task/info"
	"github.com/johnmikee/cuebert/db/shared/overdue"
)

// CheckCompliance looks at all users and removes them if they are compliant.
func (c *Shared) CheckCompliance(time.Time) {
	users, err := c.impl.CheckComplianceAll()
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("failed to get compliance users")
		return
	}

	c.log.Trace().Interface("users", users).Msg("compliance users")

	// if any of the users in the overdue are not in the users slice they should be removed
	ou, err := c.getOverdue()
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("failed to get overdue users")
		return
	}
	c.log.Trace().Interface("overdue users", ou).Msg("overdue users")

	emailMap := info.PL(users).EmailMap()

	remove := []overdue.Info{}
	for _, item := range ou {
		if _, exists := emailMap[item.UserEmail]; !exists {
			remove = append(remove, item)
		}
	}

	err = c.removeOverdue(remove)
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("failed to remove overdue users")
		return
	}

	c.log.Info().Msg("successfully removed users")
}

func (c *Shared) removeOverdue(ou []overdue.Info) error {
	conn, err := c.db.GetPool().Acquire(context.Background())
	if err != nil {
		c.log.Info().AnErr("acquiring connection", err).Msg("failed to acquire lock")
		return err
	}

	query, args, err := squirrel.Delete("overdue").Where(squirrel.Eq{"user_id": ou}).ToSql()
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("failed to build delete statement")
	}

	_, err = conn.Exec(context.Background(), query, args...)
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("failed to execute delete statement")
	}

	conn.Release()

	return nil
}

func (c *Shared) getOverdue() ([]overdue.Info, error) {
	conn, err := c.db.GetPool().Acquire(context.Background())
	if err != nil {
		c.log.Info().AnErr("acquiring connection", err).Msg("failed to acquire lock")
		return nil, err
	}

	query, args, err := squirrel.Select("*").From("overdue").ToSql()
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("failed to build select statement")
		return nil, err
	}

	rows, err := conn.Query(context.Background(), query, args...)
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("failed to execute query")
		return nil, err
	}

	ou := []overdue.Info{}

	for rows.Next() {
		var ti overdue.Info
		err = rows.Scan(
			&ti.UserID,
			&ti.UserServiceID,
			&ti.UserEmail,
			&ti.Task,
		)
		if err != nil {
			c.log.Info().AnErr("error", err).Msg("failed to scan row")
			continue
		}
		ou = append(ou, ti)
	}

	conn.Release()
	return ou, nil
}
