package common

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/johnmikee/cuebert/bot"
	"github.com/johnmikee/cuebert/cuebert/config"
	"github.com/johnmikee/cuebert/cuebert/task/shared"
	"github.com/johnmikee/cuebert/cuebert/task/update/db/devices"
	"github.com/johnmikee/cuebert/cuebert/task/update/db/tables"
	"github.com/johnmikee/cuebert/cuebert/task/update/db/tracking"
	"github.com/johnmikee/cuebert/db/shared/users"
	"github.com/johnmikee/cuebert/handlers"
	"github.com/johnmikee/cuebert/idp"
	"github.com/johnmikee/cuebert/idp/client"
	"github.com/johnmikee/cuebert/mdm"
	mdmclient "github.com/johnmikee/cuebert/mdm/client"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/slack-go/slack"
)

// Common holds the common values for the program
type Common struct {
	bot    *bot.Bot
	config *config.Config
	flags  *config.Flags
	idp    idp.Provider
	mdm    mdm.Provider
	log    logger.Logger
	tables *tables.Config
	status *handlers.StatusHandler
	shared *shared.Shared
}

// FirstMessageDelivered sets the first message delivered flag for the given slack id
func (c *Common) FirstMessageDelivered(serviceID, slackID string) error {
	// _, err := c.tables.Tracking().Update().WithOptions(
	// 	tracking.Info{
	// 		UserServiceID:      serviceID,
	// 		UserSlackID:        slackID,
	// 		FirstMessageSent:   true,
	// 		FirstMessageSentAt: helpers.UpdateTime(),
	// 	}).Parse(tracking.UserSlackID, slackID).Send()
	tableName := c.tables.Tracking().Table()
	query := fmt.Sprintf("UPDATE %s SET first_message_sent = true, first_message_sent_at = $1 WHERE user_slack_id = $2;", tableName)
	// query := "UPDATE $1 SET first_message_sent = true, first_message_sent_at = $2 WHERE user_slack_id = $3;"

	_, err := c.tables.Pool().Exec(context.Background(), query, helpers.UpdateTime(), slackID)

	return err
}

// FirstMessageSent implements shared.Caller.
func (c *Common) FirstMessageSent(serviceID string, slackID string) error {
	tableName := c.tables.Tracking().Table()
	query := fmt.Sprintf("UPDATE %s SET first_message_sent = true, first_message_sent_at = $1 WHERE user_slack_id = $2;", tableName)
	_, err := c.tables.Pool().Exec(context.Background(), query, helpers.UpdateTime(), slackID)
	return err
}

// FinalMessageSent implements shared.Caller.
func (c *Common) FinalMessageSent(serviceID string, slackID string) error {
	tableName := c.tables.Tracking().Table()
	query := fmt.Sprintf("UPDATE %s SET final_message_sent = true, final_message_sent_at = $1 WHERE user_slack_id = $2;", tableName)
	_, err := c.tables.Pool().Exec(context.Background(), query, helpers.UpdateTime(), slackID)
	return err
}

// ReminderMessage delivers the reminder message
func (c *Common) ReminderMessage(rp *shared.ReminderPayload) error {
	return c.shared.DeliverReminder(
		&shared.Reminder{
			Text:     "Hi there! Just a gentle reminder to acknowledge the previous message about updating.",
			Deadline: c.flags.Deadline,
			Cutoff:   c.flags.CutoffTime,
			User:     rp.UserSlackID,
			Task:     rp.Task,
			Title:    "Update Reminder",
			Fields: []slack.AttachmentField{
				{
					Title: "Task",
					Value: rp.Task,
				},
				{
					Title: "Update Deadline",
					Value: c.flags.Deadline + " " + c.flags.CutoffTime,
				},
			},
		},
	)
}

type Config struct {
	Bot    *bot.Bot
	Config *config.Config
	Flags  *config.Flags
	Log    logger.Logger
	Tables *tables.Config
	Status *handlers.StatusHandler
	Shared *shared.Shared
}

// New will return a new User struct
func New(u *Config) *Common {
	c := &Common{
		bot:    u.Bot,
		config: u.Config,
		idp: client.New(
			&client.IDP{
				IDP: idp.IDP(u.Flags.IDP),
				Config: idp.Config{
					Domain:        u.Config.Subdomain,
					URL:           u.Config.IDPURL,
					Token:         u.Config.IDPToken,
					Client:        nil,
					Log:           u.Log,
					IDPBlockGroup: u.Config.IDPBlockGroup,
				},
			},
		),
		mdm: mdmclient.New(
			&mdmclient.MDM{
				MDM: mdm.MDM(u.Flags.MDM),
				Config: mdm.Config{
					Domain: u.Config.Subdomain,
					MDM:    mdm.MDM(u.Flags.MDM),
					URL:    u.Config.MDMURL,
					Token:  u.Config.MDMKey,
					Client: nil,
					Log:    u.Log,
				},
			},
		),
		flags:  u.Flags,
		log:    logger.ChildLogger("common", &u.Log),
		tables: u.Tables,
		status: u.Status,
		shared: u.Shared,
	}

	return c
}

func (c *Common) LogFlags() {
	config.LogFlags(c.log, c.flags)
}

type Overlap struct {
	Serial   string
	User     string
	SlackID  string
	OS       string
	TZOffset int64
	FullName string
}

// AddAllDevices will add all devices from the MDM to the DB
func (c *Common) AddAllDevices(dev []mdm.Device) error {
	dev, err := c.listDevices(dev)
	if err != nil {
		return err
	}

	machines := devices.DI{}

	for i := range dev {
		if dev[i].Platform == "Mac" {
			di := devices.Info{
				DeviceID:     dev[i].DeviceID,
				DeviceName:   dev[i].DeviceName,
				Model:        dev[i].Model,
				SerialNumber: dev[i].SerialNumber,
				Platform:     dev[i].Platform,
				OSVersion:    dev[i].OSVersion,
				LastCheckIn:  dev[i].LastCheckIn,
				User:         dev[i].User.Email,
				UserMDMID:    dev[i].User.ID,
			}

			machines = append(machines, di)
		}
	}

	_, err = c.tables.Device().AddAllDevices(machines)
	if err != nil {
		return err
	}

	return nil
}

// AddAllUsers will add all users from the MDM to the DB
func (c *Common) AddAllUsers(dev []mdm.Device) ([]string, error) {
	res, err := c.listDevices(dev)
	if err != nil {
		return nil, err
	}

	c.log.Info().Int("count", len(res)).Msg("adding users to the database")

	deviceUserMap := map[string]string{}
	for i := range res {
		deviceUserMap[res[i].User.Email] = res[i].User.ID
	}

	us := users.UI{}

	c.log.Trace().Msg("getting users from slack...")
	su, err := c.bot.Bot().SlackClient().GetUsers()
	if err != nil {
		return nil, err
	}

	doubleCheck := []string{}

	for i := range su {
		if su[i].Profile.Email == "" {
			c.log.Trace().
				Str("name", su[i].Profile.RealName).
				Msg("skipping profile with no email")
			continue
		}
		if su[i].Deleted {
			c.log.Debug().
				Str("email", su[i].Profile.Email).
				Msg("skipping deleted account")
			continue
		}

		ui := users.Info{}
		ui.UserEmail = su[i].Profile.Email
		ui.UserLongName = su[i].Profile.RealName
		ui.UserSlackID = su[i].ID
		ui.TZOffset = int64(su[i].TZOffset)

		i, ok := deviceUserMap[ui.UserEmail]
		if !ok {
			doubleCheck = append(doubleCheck, ui.UserEmail)
			c.log.Debug().Str("skipping email", ui.UserEmail).Msg("adding user to double check slice")
			continue
		}

		ui.UserServiceID = func() int {
			c.log.Trace().Str("item", i).Msg("converting user id")
			i, err := strconv.Atoi(i)
			if err != nil {
				c.log.Err(err).Msg("could not convert user id")
			}
			return i
		}()

		us = append(us, ui)
	}

	_, err = c.tables.User().Add().All(us)
	if err != nil {
		return nil, err
	}

	return doubleCheck, err
}

func (c *Common) listDevices(devices []mdm.Device) ([]mdm.Device, error) {
	if devices != nil {
		return devices, nil
	}
	c.log.Trace().Str("mdm", c.flags.MDM).Msg("getting devices from mdm...")
	return c.mdm.ListDevices()
}

func (c *Common) MessageAcknowledged(userID string, when time.Time, which string) error {
	var err error
	switch which {
	case "first":
		c.log.Info().Str("user", userID).Msg("first ack")
		_, err = c.tables.Tracking().Update().WithOptions(
			tracking.Info{
				FirstACK:     true,
				FirstACKTime: when,
				UserSlackID:  userID,
			}).Parse(tracking.UserSlackID, userID).Send()
	case "reminder":
		c.log.Info().Str("user", userID).Msg("reminder ack")
		_, err = c.tables.Tracking().Update().WithOptions(
			tracking.Info{
				UserReminderAt:   when,
				UserReminderTime: when.Format("15:04"),
				UserReminderDate: when.Format("2006-01-02"),
				UserSlackID:      userID,
			}).Parse(tracking.UserSlackID, userID).Send()
	default:
		c.log.Info().Str("which", which).Msg("unknown ack type")
	}

	return err
}
