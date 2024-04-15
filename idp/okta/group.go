package okta

import (
	"fmt"
	"net/http"
	"time"

	"github.com/johnmikee/cuebert/pkg/helpers"
)

// GroupMembers represents the response from the Okta API containing a slice of the group members.
type GroupMembers []struct {
	ID            string    `json:"id"`
	Status        string    `json:"status"`
	Created       time.Time `json:"created"`
	Activated     time.Time `json:"activated"`
	StatusChanged time.Time `json:"statusChanged"`
	LastLogin     time.Time `json:"lastLogin"`
	LastUpdated   time.Time `json:"lastUpdated"`
	Profile       Profile   `json:"profile"`
	Links         Links     `json:"_links"`
}

// Provider represents the Okta IDP provider.
type Provider struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// Self contains reference links from Okta returned in the response.
type Self struct {
	Href string `json:"href"`
}

// Links contains reference links from Okta returned in the response.
type Links struct {
	Self Self `json:"self"`
}

var groupBase = "groups"

// GetAdminGroup returns a slice of emails for the users in the admin group.
func (o *Client) GetAdminGroup(groupID string) ([]string, error) {
	members, err := o.getGroupsMembers(groupID)
	if err != nil {
		return nil, err
	}

	memberEmails := make([]string, len(*members))
	for _, member := range *members {
		memberEmails = append(memberEmails, member.Profile.Email)
	}

	return helpers.RemoveEmpty(memberEmails), err
}

func (o *Client) GetBlockedGroup() ([]string, error) {
	members, err := o.getGroupsMembers(o.blockGroup)
	if err != nil {
		return nil, err
	}

	memberEmails := make([]string, len(*members))
	for _, member := range *members {
		memberEmails = append(memberEmails, member.Profile.Email)
	}

	return helpers.RemoveEmpty(memberEmails), err
}

func (o *Client) getGroupsMembers(groupID string) (*GroupMembers, error) {
	req, err := o.newRequest(http.MethodGet, fmt.Sprintf("%s/%s/users", groupBase, groupID), false, nil)
	if err != nil {
		return nil, err
	}

	var groupMembers GroupMembers
	_, err = o.do(req, &groupMembers)
	if err != nil {
		return nil, err
	}

	return &groupMembers, nil
}

func (o *Client) addUserToGroup(groupID, userID string) (*http.Response, error) {
	return o.toGroup(http.MethodPut, groupID, userID)
}

func (o *Client) removeUserFromGroup(groupID, userID string) (*http.Response, error) {
	return o.toGroup(http.MethodDelete, groupID, userID)
}

func (o *Client) toGroup(method, groupID, userID string) (*http.Response, error) {
	req, err := o.newRequest(method, fmt.Sprintf("%s/%s/users/%s", groupBase, groupID, userID), false, nil)
	if err != nil {
		return nil, err
	}

	return o.do(req, nil)
}
