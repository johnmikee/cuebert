package tables

import (
	"strconv"

	"github.com/johnmikee/cuebert/db/shared/users"
	"github.com/johnmikee/cuebert/idp"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/slack-go/slack"
)

// we take the users from the iDP and compare them to the users in the database.
func (c *Config) AddCheckMissing(idpUsers []idp.User, check []string, sc *slack.Client) {
	emails := make([]string, len(idpUsers))
	emailProfileMap := make(map[string]idp.User)
	for i := range idpUsers {
		emails = append(emails, idpUsers[i].Profile.Email)
		emailProfileMap[idpUsers[i].Profile.Email] = idpUsers[i]
	}

	for u := range check {
		c.log.Trace().Str("user", check[u]).Msg("checking user")
		if helpers.Contains(emails, check[u]) {
			// now we need their slackID
			su, err := sc.GetUserByEmail(check[u])
			c.log.Trace().Str("email", check[u]).Msg("getting user by email")

			if err != nil {
				c.log.Debug().AnErr("getting user by email", err).
					Str("email", check[u]).
					Send()
				continue
			}
			profile, ok := emailProfileMap[check[u]]
			if !ok {
				c.log.Trace().Msg("user not found in emailProfileMap")
				continue
			}
			_, err = c.User().Add().WithOptions(
				users.Info{
					UserServiceID: func() int {
						i, err := strconv.Atoi(profile.ID)
						if err != nil {
							c.log.Err(err).Msg("could not convert user id")
						}
						return i
					}(),
					UserEmail:    check[u],
					UserSlackID:  su.ID,
					UserLongName: su.RealName,
				}).Execute()
			if err != nil {
				c.log.Debug().AnErr("adding user", err).
					Str("email", check[u]).
					Send()
				continue
			}
		}
	}
}

// func (c *Config) updateAdmin() error {
// 	return c.Admin().Update().WithOptions(admin.Info{
// 		HasBuiltTables: true,
// 		HasStarted:     true,
// 		UpdatedAt:      helpers.UpdateTime(),
// 	}).Send()
// }
