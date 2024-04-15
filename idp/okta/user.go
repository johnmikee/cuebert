package okta

import (
	"fmt"
	"net/http"
	"time"

	"github.com/johnmikee/cuebert/idp"
)

// UserResponse is a slice of User responses returned when querying the /users endpoint
type UserResponse []User

// User holds information on the user returned when querying the /users endpoint
type User struct {
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

// GetAllUsers returns a slice of users formatted for the IDP interface.
func (c *Client) GetAllUsers() (idp.UI, error) {
	return c.getAllUsers()
}

// GetUser returns a single user formatted for the IDP interface.
func (c *Client) GetUser(user string) (*idp.User, error) {
	return c.getUser(user)
}

func (o *Client) user(user string) (*User, error) {
	url := fmt.Sprintf("users/%s", user)

	req, err := o.newRequest(http.MethodGet, url, false, nil)
	if err != nil {
		return nil, err
	}

	var userRes User
	_, err = o.do(req, &userRes)
	if err != nil {
		return nil, err
	}

	return &userRes, nil
}

func (o *Client) getUser(user string) (*idp.User, error) {
	url := fmt.Sprintf("users/%s", user)

	req, err := o.newRequest(http.MethodGet, url, false, nil)
	if err != nil {
		return nil, err
	}

	var userRes User
	_, err = o.do(req, &userRes)
	if err != nil {
		return nil, err
	}

	iu := idp.User{
		ID: userRes.ID,
		Profile: idp.Profile{
			LastName:  userRes.Profile.LastName,
			Manager:   userRes.Profile.Manager,
			ManagerID: userRes.Profile.ManagerID,
			Title:     userRes.Profile.Title,
			Login:     userRes.Profile.Login,
			FirstName: userRes.Profile.FirstName,
			UserType:  user,
			Email:     userRes.Profile.Email,
			StartDate: userRes.Profile.StartDate,
		},
	}

	return &iu, nil
}

func (o *Client) getActive(url string, override bool) (*http.Response, UserResponse, error) {
	req, err := o.newRequest(http.MethodGet, url, override, nil)
	if err != nil {
		return nil, nil, err
	}

	var ur UserResponse
	resp, err := o.do(req, &ur)
	if err != nil {
		return nil, nil, err
	}

	return resp, ur, nil
}

func (o *Client) getAllUsers() (idp.UI, error) {
	results := []idp.User{}

	var override bool
	url := "users?filter=status+eq+%22ACTIVE%22"
	for {
		resp, ur, err := o.getActive(url, override)
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

		url = link
		override = true
		o.log.Trace().Msg("checking next link..")
	}

	return results, nil
}
