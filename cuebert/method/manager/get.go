package manager

import (
	"github.com/johnmikee/cuebert/db/bot"
	"github.com/johnmikee/cuebert/idp"
	"github.com/johnmikee/cuebert/pkg/helpers"
)

// getManagers determines, if possible, the managers for the users that are missing them.
func (m *Manager) getManagers(mm []MissingManager, ur []idp.User) (bot.BR, []MissingManager, error) {
	du, err := m.tables.GetAllUsers()
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

	update := bot.BR{}
	MissingManagers := []MissingManager{}

	for i := range ur {
		_, ok := helpers.ContainsPosition(userMissingManager, ur[i].Profile.Email)
		if !ok {
			m.log.Trace().
				Str("user", ur[i].Profile.Email).
				Msg("already associated manager")
			continue
		}
		managerEmail := ur[i].Profile.ManagerID
		managerID, ok := emailIDMap[managerEmail]
		userID := emailIDMap[ur[i].Profile.Email]
		if ok {
			m.log.Trace().
				Str("user", ur[i].Profile.Email).
				Str("manager", managerID).
				Msg("found manager")

			update = append(update, bot.Info{
				UserEmail:      ur[i].Profile.Email,
				ManagerSlackID: managerID,
				SlackID:        userID,
			})
		} else {
			m.log.Debug().
				Str("user", ur[i].Profile.Email).
				Msg("no manager found")

			MissingManagers = append(MissingManagers, MissingManager{
				userEmail: ur[i].Profile.Email,
				user:      ur[i].Profile.FirstName + " " + ur[i].Profile.LastName,
			})
		}
	}
	return update, MissingManagers, nil
}

// getMissingManagers returns a list of users that do not have a manager value sem.
func (m *Manager) getMissingManagers() ([]MissingManager, error) {
	m.log.Debug().Msg("getting users missing managers from db")
	bt, err := m.tables.NoManager()
	if err != nil {
		return nil, err
	}

	mm := []MissingManager{}
	// see who is missing a manager in the table
	for i := range bt {
		if bt[i].ManagerSlackID == "" {
			mm = append(mm, MissingManager{
				userEmail: bt[i].UserEmail,
				user:      bt[i].FullName,
			})
		}
	}

	return mm, nil
}
