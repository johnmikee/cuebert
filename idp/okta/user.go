package okta

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/johnmikee/cuebert/idp"
)

// UserResponse holds information on the user returned when querying the /users endpoint
type UserResponse struct {
	ID              string    `json:"id"`
	Status          string    `json:"status"`
	Created         time.Time `json:"created"`
	Activated       time.Time `json:"activated"`
	StatusChanged   time.Time `json:"statusChanged"`
	LastLogin       time.Time `json:"lastLogin"`
	LastUpdated     time.Time `json:"lastUpdated"`
	PasswordChanged time.Time `json:"passwordChanged"`
	Profile         Profile   `json:"profile"`
}

// Profiles holds information on the users profile
type Profile struct {
	LastName    string `json:"lastName"`
	Manager     string `json:"manager"`
	SecondEmail string `json:"secondEmail"`
	ManagerID   string `json:"managerId"`
	Title       string `json:"title"`
	Login       string `json:"login"`
	FirstName   string `json:"firstName"`
	UserType    string `json:"userType"`
	Department  string `json:"department"`
	StartDate   string `json:"startDate"`
	Email       string `json:"email"`
}

// UserOpts holds the three options for user name when querying the /users endpoint
type UserOpts struct {
	ID             string
	Login          string
	LoginShortname string
}

func (o *Client) userOptSorter(u *UserOpts) (string, bool) {
	// we only need one of these passed. if the count is over 1 we will
	// go in order of id, shortname, and then login as the return and log
	// a warning.
	count := 0
	var user string

	if u.ID != "" {
		user = u.ID
		count++
	}
	if u.LoginShortname != "" {
		user = userNameChecker(u.LoginShortname, o.domain)
		count++
	}
	if u.Login != "" {
		user = userNameChecker(u.Login, o.domain)
		count++
	}

	if count > 1 {
		o.log.Warn().Msg("only one option should be passed to GetUser from UserOpts.")
		o.log.Debug().Msgf(
			"the following values were passed - ID: %s, Login: %s, LoginShortname: %s", u.ID, u.Login, u.LoginShortname)
	}

	if user == "" {
		return user, false
	}

	return user, true
}

// GetUser will return a UserResponse body for the user queried against.
// user:
//   - when querying a user you must pass either the id, login, or loginshortname.
//     this function takes all three options and will be sorted by the userOptSorter
//     function.
func (o *Client) GetUser(user *UserOpts) (*UserResponse, *http.Response, error) {
	u, ok := o.userOptSorter(user)
	if !ok {
		msg := "could not generate user query"
		return nil, nil, errors.New(msg)
	}

	url := fmt.Sprintf("users/%s", u)

	req, err := o.newRequest(http.MethodGet, url, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	var userRes UserResponse
	resp, err := o.do(req, &userRes)
	if err != nil {
		return nil, nil, err
	}

	return &userRes, resp, nil
}

func (o *Client) getActive(u *urlOverride) (*http.Response, []UserResponse, error) {
	url := `users?filter=status+eq+%22ACTIVE%22`

	if u.override {
		url = u.url
	}
	req, err := o.newRequest(http.MethodGet, url, u, nil)

	if err != nil {
		return nil, nil, err
	}

	var ur []UserResponse
	resp, err := o.do(req, &ur)
	if err != nil {
		return resp, nil, err
	}

	return resp, ur, nil
}

// GetAllOktaUsers will return an array of all active Okta users
func (o *Client) getAllUsers() ([]idp.User, error) {
	results := []idp.User{}

	u := &urlOverride{override: false}
	for {
		resp, ur, err := o.getActive(u)
		if err != nil {
			return nil, err
		}

		for i := range ur {
			if ur[i].Profile.UserType != "Service Account" {
				results = append(results,
					idp.User{
						ID:        ur[i].ID,
						Status:    ur[i].Status,
						Activated: ur[i].Activated,
						Profile: idp.Profile{
							LastName:   ur[i].Profile.LastName,
							Manager:    ur[i].Profile.Manager,
							ManagerID:  ur[i].Profile.ManagerID,
							Title:      ur[i].Profile.Title,
							Login:      ur[i].Profile.Login,
							FirstName:  ur[i].Profile.FirstName,
							UserType:   ur[i].Profile.UserType,
							Department: ur[i].Profile.Department,
							Email:      ur[i].Profile.Email,
						},
					},
				)
			}
		}

		value := resp.Header["Link"]

		link := linkSorter(value)
		if link == "" {
			o.log.Trace().Msg("no more responses from okta")
			break
		}
		u.override = true
		u.url = link
		o.log.Trace().Msg("checking next link..")

	}

	return results, nil
}
