package common

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/johnmikee/cuebert/cuebert/task/info"
	"github.com/johnmikee/cuebert/cuebert/task/update/db/tracking"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/pkg/errors"
)

// TableBuilder is a struct to hold the data needed to build the tables.
type TableBuilder struct {
	Users []info.Users
	Tasks []info.Tasks
}

// TableBuilder implements task.Rabbit.
func (c *Common) TableBuilder(t *TableBuilder) ([]string, error) {
	c.log.Info().Msg("initializing tables...")

	dev, err := c.listDevices(nil)
	if err != nil {
		return nil, err
	}
	check, err := c.AddAllUsers(dev)
	if err != nil {
		c.log.Err(err).Msg("could not add user table")
	}
	err = c.AddAllDevices(dev)
	if err != nil {
		c.log.Err(err).Msg("could not add device table")
	}

	// build bot_results table. we need this built before we can do anything else
	c.buildTrackingTable(c.flags.RequiredVers)

	return check, err
}

// buildTrackingTable builds the tracking table which tracks the users that need to be reminded to update their devices.
func (c *Common) buildTrackingTable(reqVers string) {
	updates := tracking.TI{}

	br, err := c.deviceUserOverlap()
	if err != nil {
		c.log.Err(err).Msg("could not build tracking table")
		return
	}

	for i := range br {
		c.log.Trace().Str("os", br[i].OS).Str("req", reqVers).Msg("checking device")
		ok, err := helpers.CompareOSVer(br[i].OS, reqVers)
		if err != nil {
			c.log.Err(err).Msg("could not compare os versions")
		}

		if !ok {
			c.log.Debug().
				Str("serial", br[i].Serial).
				Str("os", br[i].OS).
				Msg("needs update")

			u := tracking.Info{
				UserSlackID:  br[i].SlackID,
				UserEmail:    br[i].User,
				UserFullName: br[i].FullName,
				SerialNumber: br[i].Serial,
				TZOffset:     br[i].TZOffset,
			}
			updates = append(updates, u)
		}
	}
	_, err = c.tables.Tracking().Add().All(updates)
	if err != nil {
		c.log.Err(err).Msg("could not build bot results table")
	}
}

func DUOverlapQuery() (string, []interface{}, error) {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("device_user, user_long_name, serial_number, user_slack_id, os_version, tz_offset").
		From("updates_os_devices d").
		Join("updates_os_users ON (user_email = d.device_user)").
		Where(sq.And{sq.NotEq{"device_user": ""}, sq.NotEq{"user_slack_id": ""}}).
		ToSql()
}

// DeviceUserOverlap returns a list of devices that have a user name and a slack id.
func (c *Common) deviceUserOverlap() ([]Overlap, error) {
	query, args, err := DUOverlapQuery()
	if err != nil {
		return nil, errors.Wrap(err, "device user overlap query failed")
	}
	rows, err := c.tables.GetConnection().Query(context.Background(), query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "tracking table query failed")
	}

	resp := []Overlap{}
	for rows.Next() {
		var br Overlap

		err = rows.Scan(
			&br.User,
			&br.FullName,
			&br.Serial,
			&br.SlackID,
			&br.OS,
			&br.TZOffset,
		)
		if err != nil {
			return nil, errors.Wrap(err, "rows scan failed")
		}
		resp = append(resp, br)
	}

	return resp, nil
}
