package user

import (
	"fmt"
	"strings"

	"github.com/johnmikee/cuebert/db"
	"github.com/johnmikee/cuebert/db/users"
	"github.com/johnmikee/cuebert/mdm"
	"github.com/slack-go/slack"

	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// User is the internal struct for the user package
type User struct {
	sc     *slack.Client
	db     *db.DB
	client mdm.Provider
	log    logger.Logger
}

// UserConfig is the configuration for the User struct.
// This includes the mdm info to query devices
type Config struct {
	Slack  *slack.Client
	Client mdm.Provider
	DB     *db.DB
	Log    *logger.Logger
}

// New will return a new User struct
func New(u *Config) *User {
	return &User{
		sc:     u.Slack,
		db:     u.DB,
		client: u.Client,
		log:    logger.ChildLogger("user", u.Log),
	}
}

// AddAllUsers will add all users from the MDM to the DB
func (c *User) AddAllUsers() ([]string, error) {
	res, err := c.client.ListDevices()
	if err != nil {
		return nil, err
	}

	// we need the MDM ID for the table. Instead of nesting two for loops
	// we add the email and ID and use a search on the slice and split the
	// string if a match is found on the email while iterating the slack response
	mu := []string{}
	for i := range res {
		mu = append(
			mu,
			fmt.Sprintf("%s::%v", res[i].User.Email, res[i].User.ID),
		)
	}
	us := users.UI{}

	su, err := c.sc.GetUsers()
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

		i, ok := helpers.ContainsPosition(mu, ui.UserEmail)
		if !ok {
			doubleCheck = append(doubleCheck, ui.UserEmail)
			c.log.Debug().Str("skipping email", ui.UserEmail).Msg("adding user to double check slice")
			continue
		}
		item := mu[i]
		resp := strings.Split(item, "::")
		if len(resp) >= 1 {
			ui.MDMID = resp[1]
		}
		us = append(us, ui)
	}

	db := users.User(c.db, &c.log)

	_, err = db.AddAllUsers(us)
	if err != nil {
		return nil, err
	}

	return doubleCheck, err
}

// GetMDMUsers will return all users from the MDM
func (c *User) GetMDMUsers(opts *mdm.QueryOpts) ([]mdm.User, error) {
	var res mdm.DeviceResults
	var err error

	if opts != nil {
		res, err = c.withQuery(opts)
	} else {
		res, err = c.all()
	}

	if err != nil {
		return nil, err
	}

	mu := []mdm.User{}
	for i := range res {
		mu = append(mu, res[i].User)
	}

	return mu, nil
}

func (c *User) all() (mdm.DeviceResults, error) {
	return c.client.ListDevices()
}

func (c *User) withQuery(opts *mdm.QueryOpts) (mdm.DeviceResults, error) {
	return c.client.QueryDevices(opts)
}
