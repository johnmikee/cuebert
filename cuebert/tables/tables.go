package tables

import (
	"context"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/johnmikee/cuebert/cuebert/device"
	"github.com/johnmikee/cuebert/cuebert/user"
	"github.com/johnmikee/cuebert/db"
	"github.com/johnmikee/cuebert/db/bot"
	"github.com/johnmikee/cuebert/idp"
	"github.com/johnmikee/cuebert/mdm"

	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/slack-go/slack"
)

type Check struct {
	Serial string
	OS     string
}

type DBClient struct {
	DB     *db.DB
	Log    logger.Logger
	Device *device.Device
	User   *user.User
}

func (c *Config) InitTables(reqVers string) ([]string, error) {
	check, err := c.users.AddAllUsers()
	if err != nil {
		c.log.Err(err).Msg("could not add user table")
	}

	err = c.devices.AddAllDevices()
	if err != nil {
		c.log.Err(err).Msg("could not add device table")
	}

	// build bot_results table. we need this built before we can do anything else
	c.BuildBotResTable(reqVers)

	return check, err
}

// buildBotResTable builds the bot results table. this is used to track
// the users that need to be reminded to update their devices.
func (c *Config) BuildBotResTable(reqVers string) {
	updates := bot.BR{}

	br, err := c.DeviceUserOverlap()
	if err != nil {
		c.log.Err(err).Msg("could not build bot results table")
		return
	}

	for i := range br {
		ok, err := helpers.CompareOSVer(br[i].OS, reqVers)
		if err != nil {
			c.log.Err(err).Msg("could not compare os versions")
		}

		if !ok {
			c.log.Debug().
				Str("serial", br[i].Serial).
				Str("os", br[i].OS).
				Msg("needs update")

			u := bot.Info{
				SlackID:      br[i].SlackID,
				UserEmail:    br[i].User,
				FullName:     br[i].FullName,
				SerialNumber: br[i].Serial,
				FirstACKTime: time.Time{},
				DelayAt:      time.Time{},
				TZOffset:     br[i].TZOffset,
			}
			updates = append(updates, u)
		}
	}

	err = c.BatchAddBotInfo(updates)
	if err != nil {
		c.log.Err(err).Msg("could not build bot results table")
	}
}

// we take the users from the iDP and compare them to the users in the database.
func (c *Config) AddCheckMissing(idpUsers []idp.User, users []string, sc *slack.Client) {
	emails := make([]string, len(idpUsers))
	emailProfileMap := make(map[string]idp.User)
	for i := range idpUsers {
		emails = append(emails, idpUsers[i].Profile.Email)
		emailProfileMap[idpUsers[i].Profile.Email] = idpUsers[i]
	}

	for u := range users {
		c.log.Trace().Str("user", users[u]).Msg("checking user")
		if helpers.Contains(emails, users[u]) {
			// now we need their slackID
			su, err := sc.GetUserByEmail(users[u])
			c.log.Trace().Str("email", users[u]).Msg("getting user by email")

			if err != nil {
				c.log.Debug().AnErr("getting user by email", err).
					Str("email", users[u]).
					Send()
				continue
			}
			profile, ok := emailProfileMap[users[u]]
			if !ok {
				c.log.Trace().Msg("user not found in emailProfileMap")
				continue
			}
			_, err = c.AddUser().
				ID(profile.ID).
				Email(users[u]).
				Slack(su.ID).
				LongName(su.RealName).
				Execute()
			if err != nil {
				c.log.Debug().AnErr("adding user", err).
					Str("email", users[u]).
					Send()
				continue
			}
		}
	}
}

// deleteTables deletes the tables specified. defaults to all tables.
func (c *Config) DeleteTables(tables string) error {
	names := strings.Split(tables, ",")

	return c.delete(names)
}

// GatherDiffDevicesDB gets all the devices from the db and checks which
// ones do not have the required os version.
func (c *Config) GatherDiffDevicesDB() ([]string, []Check, error) {
	dbd, err := c.GetAllDevices()
	if err != nil {
		c.log.Debug().
			AnErr("getting db devices", err).
			Send()
		return nil, nil, err
	}

	c.log.Trace().Msg("got db devices")

	ds := []string{}
	versCheck := []Check{}

	for i := range dbd {
		ds = append(ds, dbd[i].SerialNumber)
		versCheck = append(versCheck, Check{
			Serial: dbd[i].SerialNumber,
			OS:     dbd[i].OSVersion,
		})
	}

	return ds, versCheck, nil
}

// returns a bool indicating if the serial number is in the exclusions table and a bool for the approval status
func (c *Config) IsExcluded(serial string) (bool, bool) {
	ex, err := c.SerialExcluded(serial)
	if err != nil {
		c.log.Err(err).Msg("could not check if device was in exclusions table")
		return false, false
	}

	if len(ex) == 0 {
		return false, false
	}

	if ex[0].SerialNumber == serial {
		return true, ex[0].Approved
	}

	return false, false
}

// UpdateReminderTime updates the reminder time in the db.
func (c *Config) UpdateReminderTime(dv, tv, sid string) {
	dateString := fmt.Sprintf("%s %s", dv, tv)
	ts, err := time.Parse("2006-01-02 15:04", dateString)
	if err != nil {
		c.log.Err(err).Msg("could not compose time string")
	}

	err = c.SetReminder(sid, ts.UTC(), dv, tv)
	if err != nil {
		c.log.Err(err).Msg("could not record the first ack time")
	}
}

func (c *Config) delete(tables []string) error {
	for _, t := range tables {
		if !helpers.Contains(db.CueTables, t) {
			return fmt.Errorf("%s is not a table", t)
		}

		query, args, err := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
			Delete(t).
			ToSql()

		c.log.Trace().
			Str("query", query).
			Str("args", fmt.Sprintf("%v", args)).
			Str("table", t).
			Msg("delete query")

		if err != nil {
			return fmt.Errorf("building delete query for %s failed: %s", t, err)
		}

		_, err = c.db.Exec(context.Background(), query, args...)
		if err != nil {
			return fmt.Errorf("deleting table %s failed: %s", t, err)
		}
	}

	return nil
}

// checkStaleDevices checks for devices that no longer need to be in the DB.
// If the OS on the device is greater than the required version, it will be removed.
func CheckStaleDevices(reqVers string, versCheck []Check, md mdm.DeviceResults) []string {
	remove := []string{}

	for i := range md {
		for _, x := range versCheck {
			if md[i].SerialNumber == x.Serial {
				ok, _ := helpers.CompareOSVer(x.OS, md[i].OSVersion)
				// if the MDM version is greater than the DB version
				// double check that its not above the required
				if ok {
					doubleCheck, _ := helpers.CompareOSVer(reqVers, x.OS)
					if !doubleCheck {
						remove = append(remove, x.Serial)
					}
				} else {
					ok, _ := helpers.CompareOSVer(reqVers, x.OS)
					if ok {
						remove = append(remove, x.Serial)
					}
				}
			}
		}
	}

	return remove
}
