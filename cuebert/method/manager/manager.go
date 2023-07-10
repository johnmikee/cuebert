package manager

import (
	"time"

	"github.com/johnmikee/cuebert/cuebert/bot"
	"github.com/johnmikee/cuebert/cuebert/handlers"
	"github.com/johnmikee/cuebert/cuebert/method"
	"github.com/johnmikee/cuebert/cuebert/tables"

	"github.com/johnmikee/cuebert/idp"
	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/slack-go/slack"
)

type Manager struct {
	log           logger.Logger
	tables        *tables.Config
	bot           *bot.Bot
	idp           idp.Provider
	mdm           mdm.Provider
	cfg           *Cfg
	sc            *slack.Client
	statusHandler *handlers.StatusHandler
}

// Setup implements method.Actions.
func (m *Manager) Setup(method method.Config) {
	m.log = method.Log
	m.tables = method.Tables
	m.bot = method.Bot
	m.idp = method.IDP
	m.mdm = method.MDM
	m.sc = method.SlackClient
	m.statusHandler = method.StatusHandler
	m.cfg = WithOptions(
		WithCutoffTime(method.CutoffTime),
		WithDeadline(method.Deadline),
		WithRequiredVers(method.RequiredVers),
		WithSlackAlertChannel(method.SlackAlertChannel),
		WithTesting(method.Testing),
		WithTestingUsers(method.TestingUsers),
		WithPollInterval(method.PollInterval),
	)
}

// PostInit implements method.Actions.
func (m *Manager) PostCheck(sa []string) {
	m.associateUserManager(sa)
}

// Poll implements method.Actions.
func (m *Manager) Poll(t time.Time) {
	m.log.Trace().Msg("not implemented")
}

// FirstMessage implements method.Actions.
func (*Manager) FirstMessage() string {
	return firstMessage(helpers.GetReminderDay())
}

func (m *Manager) TableAssociations(sa []string) {
	m.associateUserManager(sa)
}

type MissingManager struct {
	user      string
	userEmail string
}

const (
	GroupDM                    = "group_dm"
	ManagerMessageSentAtPicker = "manager_message_sent_at_picker"
	ManagerMessageSentAt       = "manager_message_sent_at"
	ManagerMessageSentOption   = "manager_message_sent_option"
	ManagerSlackID             = "manager_slack_id"
	UsersSelect                = "users_select"
	UpdateUser                 = "update_user"
	UserUpdateModal            = "user_update_modal"
)

func New(c *Config) *Manager {
	return &Manager{
		log:           c.Log,
		tables:        c.DB,
		bot:           c.Bot,
		idp:           c.IDP,
		mdm:           c.MDM,
		cfg:           &Cfg{},
		sc:            c.Bot.Client(),
		statusHandler: c.Handler,
	}
}
func (m *Manager) buildAssociation(idpUsers []idp.User, check []string) ([]MissingManager, error) {
	m.tables.AddCheckMissing(idpUsers, check, m.sc)

	mm, err := m.getMissingManagers()
	if err != nil {
		m.log.Err(err).
			Msg("getting missing managers from db")
	}

	update, missing, err := m.getManagers(mm, idpUsers)
	if err != nil {
		m.log.Err(err).
			Msg("getting all users managers")
		return missing, err
	}

	err = m.updateManagerVals(update)

	return missing, err
}

// check is the list of users who were not added to the db. this means they
// did not exist in the MDM. a likely cause is that the user has a non-macOS
// machine ala windows.
//
// we take a second pass on these and attach it directly to the user.
func (m *Manager) associateUserManager(check []string) {
	m.log.Info().Msg("checking manager association..")

	ur, err := m.idp.GetAllUsers()
	if err != nil {
		m.log.Trace().
			AnErr("error", err).
			Msg("getting all idp users")
		return
	}

	for i := range ur {
		m.log.Trace().
			Str("user", ur[i].Profile.Email).
			Msg("checking user")
	}

	missing, err := m.buildAssociation(ur, check)
	if err != nil {
		m.log.Err(err).
			Msg("associating managers to users")
		return
	}
	m.log.Trace().Msg("alerting if no manager")

	// if we have missing managers we need to alert

	// that is, if we opted to do so.
	m.alertIfNoManager(m.cfg.slackAlertChannel,
		[]ManagerAlert{
			{
				info: missing,
				msg:  "the db",
			},
		},
	)
}
