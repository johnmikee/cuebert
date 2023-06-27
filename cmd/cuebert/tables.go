package main

import (
	"fmt"

	"strings"
	"time"

	"github.com/johnmikee/cuebert/db/bot"
	"github.com/johnmikee/cuebert/idp"
	"github.com/johnmikee/cuebert/internal/db"
	"github.com/johnmikee/cuebert/internal/device"
	"github.com/johnmikee/cuebert/internal/user"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/johnmikee/cuebert/pkg/visual"
	"github.com/slack-go/slack"
)

// Tables is used to facilitate actions on the database.
type Tables struct {
	db      *db.Config
	log     logger.Logger
	devices *device.Device
	users   *user.User
}

func (t *Tables) initTables(reqVers string) ([]string, error) {
	check, err := t.addAllUsers()
	if err != nil {
		t.log.Err(err).Msg("could not add user table")
	}

	err = t.addAllDevices()
	if err != nil {
		t.log.Err(err).Msg("could not add device table")
	}

	// build bot_results table. we need this built before we can do anything else
	t.buildBotResTable(reqVers)

	return check, err
}

func (t *Tables) buildAssociation(
	idpUsers []idp.User,
	check []string,
	sc *slack.Client,
) ([]missingManager, error) {

	t.addCheckMissing(idpUsers, check, sc)

	mm, err := t.getMissingManagers()
	if err != nil {
		t.log.Err(err).
			Msg("getting missing managers from db")
	}

	update, missing, err := t.getManagers(mm, idpUsers)
	if err != nil {
		t.log.Err(err).
			Msg("getting all users managers")
		return missing, err
	}

	err = t.updateManagerVals(update)

	return missing, err
}

// buildBotResTable builds the bot results table. this is used to track
// the users that need to be reminded to update their devices.
func (t *Tables) buildBotResTable(reqVers string) {
	updates := []bot.BotResInfo{}

	br, err := t.db.DeviceUserOverlap()
	if err != nil {
		t.log.Err(err).Msg("could not build bot results table")
		return
	}

	for i := range br {
		ok, err := helpers.CompareOSVer(br[i].OS, reqVers)
		if err != nil {
			t.log.Err(err).Msg("could not compare os versions")
		}

		if !ok {
			t.log.Debug().
				Str("serial", br[i].Serial).
				Str("os", br[i].OS).
				Msg("needs update")

			u := bot.BotResInfo{
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

	err = t.db.BatchAddBotInfo(updates)
	if err != nil {
		t.log.Err(err).Msg("could not build bot results table")
	}
}

// we take the users from the iDP and compare them to the users in the database.
func (t *Tables) addCheckMissing(idpUsers []idp.User, users []string, sc *slack.Client) {
	emails := make([]string, len(idpUsers))
	emailProfileMap := make(map[string]idp.User)
	for i := range idpUsers {
		emails = append(emails, idpUsers[i].Profile.Email)
		emailProfileMap[idpUsers[i].Profile.Email] = idpUsers[i]
	}

	for u := range users {
		t.log.Trace().Str("user", users[u]).Msg("checking user")
		if helpers.Contains(emails, users[u]) {
			// now we need their slackID
			su, err := sc.GetUserByEmail(users[u])
			t.log.Trace().Str("email", users[u]).Msg("getting user by email")

			if err != nil {
				t.log.Debug().AnErr("getting user by email", err).
					Str("email", users[u]).
					Send()
				continue
			}
			profile, ok := emailProfileMap[users[u]]
			if !ok {
				t.log.Trace().Msg("user not found in emailProfileMap")
				continue
			}
			_, err = t.db.AddUser().
				ID(profile.ID).
				Email(users[u]).
				Slack(su.ID).
				LongName(su.RealName).
				Execute()
			if err != nil {
				t.log.Debug().AnErr("adding user", err).
					Str("email", users[u]).
					Send()
				continue
			}
		}
	}
}

// addAllDevices adds all devices from MDM to the database.
func (t *Tables) addAllDevices() error {
	return t.devices.AddAllDevices()
}

// addAllUsers adds all users from the iDP to the database.
func (t *Tables) addAllUsers() ([]string, error) {
	return t.users.AddAllUsers()
}

// buildOSReport builds a report of the number of devices by OS.
func (t *Tables) buildOSReport() (*visual.PieChartOption, error) {
	br, err := t.db.GetAllDevices()
	if err != nil {
		return nil, err
	}

	v := &visual.PieChartOption{}
	rep := make(map[string]int)

	for i := range br {
		if rep[br[i].OSVersion] == 0 {
			rep[br[i].OSVersion] = 1
		} else {
			rep[br[i].OSVersion]++
		}
	}

	for k, val := range rep {
		f := float64(val)
		v.ValueList = append(v.ValueList, f)
		v.XAxis = append(v.XAxis, k)
	}

	v.Query = "OSVersion"
	v.Text = "OS Version"

	return v, nil
}

// Report is used to determine which report to build.
type Report string

const (
	FirstMessageSent   Report = "first"
	ManagerMessageSent Report = "manager"
	ReminderRequested  Report = "reminderRequested"
)

// buildSentReport builds a report of the number of users that have been sent a message.
func (t *Tables) buildSentReport(which Report) (*visual.PieChartOption, error) {
	br, err := t.db.GetBotTableInfo()
	if err != nil {
		t.log.Debug().AnErr("getting br", err).
			Send()
		return nil, err
	}

	v := &visual.PieChartOption{}

	var sent, notSent int

	switch which {
	case FirstMessageSent:
		sent, notSent = countSentStatus(br, FirstMessageSent)
		v.Query = "FirstMessageSent"
		v.Text = "First Message Sent"
	case ManagerMessageSent:
		sent, notSent = countSentStatus(br, ManagerMessageSent)
		v.Query = "ManagerMessageSent"
		v.Text = "Manager Message Sent"
	case ReminderRequested:
		sent, notSent = countSentStatus(br, ReminderRequested)
		v.Query = "RequestedReminder"
		v.Text = "Requested Reminder"
	}

	v.ValueList = append(v.ValueList, float64(sent), float64(notSent))
	v.XAxis = append(v.XAxis, "Sent", "Not Sent")

	return v, nil
}

// deleteTables deletes the tables specified. defaults to all tables.
func (t *Tables) deleteTables(tables string) error {
	names := strings.Split(tables, ",")

	return t.db.Delete(names)
}

// getManagers determines, if possible, the managers for the users that are missing them.
func (t *Tables) getManagers(mm []missingManager, ur []idp.User) ([]bot.BotResInfo, []missingManager, error) {
	du, err := t.db.GetAllUsers()
	if err != nil {
		return nil, nil, err
	}

	userMissingManager := []string{}
	for i := range mm {
		userMissingManager = append(userMissingManager, mm[i].userEmail)
	}

	emailIDMap := make(map[string]string)
	for _, d := range du {
		emailIDMap[d.UserEmail] = d.UserSlackID
	}

	update := []bot.BotResInfo{}
	missingManagers := []missingManager{}

	for i := range ur {
		m, ok := helpers.ContainsPosition(userMissingManager, ur[i].Profile.Email)
		if !ok {
			t.log.Trace().
				Str("user", ur[i].Profile.Email).
				Str("manager", ur[m].Profile.Email).
				Msg("already associated manager")
			continue
		}
		managerEmail := ur[i].Profile.ManagerID
		managerID, ok := emailIDMap[managerEmail]
		userID := emailIDMap[ur[i].Profile.Email]
		if ok {
			t.log.Trace().
				Str("user", ur[i].Profile.Email).
				Str("manager", managerID).
				Msg("found manager")

			update = append(update, bot.BotResInfo{
				UserEmail:      ur[i].Profile.Email,
				ManagerSlackID: managerID,
				SlackID:        userID,
			})
		} else {
			t.log.Debug().
				Str("user", ur[i].Profile.Email).
				Msg("no manager found")

			missingManagers = append(missingManagers, missingManager{
				userEmail: ur[i].Profile.Email,
				user:      ur[i].Profile.FirstName + " " + ur[i].Profile.LastName,
			})
		}
	}
	return update, missingManagers, nil
}

// getMissingManagers returns a list of users that do not have a manager value set.
func (t *Tables) getMissingManagers() ([]missingManager, error) {
	t.log.Debug().Msg("getting users missing managers from db")
	bt, err := t.db.NoManager()
	if err != nil {
		return nil, err
	}

	mm := []missingManager{}
	// see who is missing a manager in the table
	for i := range bt {
		if bt[i].ManagerSlackID == "" {
			mm = append(mm, missingManager{
				userEmail: bt[i].UserEmail,
				user:      bt[i].FullName,
			})
		}
	}

	return mm, nil
}

// there may be more than one serial number. turn them into a string for the message
func (t *Tables) exceptionSerials(slackID string) []string {
	resp, err := t.db.GetUsersSerialsBot(slackID)

	if err != nil {
		t.log.Error().Err(err).Msg("could not get serials")
		return nil
	}

	serials := []string{}
	for i := range resp {
		serials = append(serials, resp[i].SerialNumber)
	}

	return serials
}

// gatherDiffDevicesDB gets all the devices from the db and checks which
// ones do not have the required os version.
func (t *Tables) gatherDiffDevicesDB() ([]string, []check, error) {
	dbd, err := t.db.GetAllDevices()
	if err != nil {
		t.log.Debug().
			AnErr("getting db devices", err).
			Send()
		return nil, nil, err
	}

	t.log.Trace().Msg("got db devices")

	ds := []string{}
	versCheck := []check{}

	for i := range dbd {
		ds = append(ds, dbd[i].SerialNumber)
		versCheck = append(versCheck, check{
			serial: dbd[i].SerialNumber,
			os:     dbd[i].OSVersion,
		})
	}

	return ds, versCheck, nil
}

// returns a bool indicating if the serial number is in the exclusions table and a bool for the approval status
func (t *Tables) isExcluded(serial string) (bool, bool) {
	ex, err := t.db.SerialExcluded(serial)
	if err != nil {
		t.log.Err(err).Msg("could not check if device was in exclusions table")
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

// updateManagerVals updates the manager values in the db.
func (t *Tables) updateManagerVals(update []bot.BotResInfo) error {
	t.log.Debug().Msg("updating manager values")

	for i := range update {
		t.log.Debug().
			Str("user", update[i].UserEmail).
			Str("slack", update[i].SlackID).
			Str("manager", update[i].ManagerSlackID).
			Msg("updating manager value")
		err := t.db.AddManagerID(update[i].SlackID, update[i].UserEmail, update[i].ManagerSlackID)
		if err != nil {
			return err
		}
	}

	return nil
}

// updateReminderTime updates the reminder time in the db.
func (t *Tables) updateReminderTime(dv, tv, sid string) {
	dateString := fmt.Sprintf("%s %s", dv, tv)
	ts, err := time.Parse("2006-01-02 15:04", dateString)
	if err != nil {
		t.log.Err(err).Msg("could not compose time string")
	}

	err = t.db.SetReminder(sid, ts.UTC(), dv, tv)
	if err != nil {
		t.log.Err(err).Msg("could not record the first ack time")
	}
}
