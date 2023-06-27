package okta

import (
	"fmt"
	"net/http"
	"time"

	"github.com/johnmikee/cuebert/pkg/helpers"
)

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

type Provider struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type Self struct {
	Href string `json:"href"`
}

type Links struct {
	Self Self `json:"self"`
}

var groupBase = "groups"

func (o *Client) GetAdminGroup(groupID string) ([]string, error) {
	members, _, err := o.getGroupsMembers(groupID)
	if err != nil {
		return nil, err
	}

	memberEmails := make([]string, len(*members))
	for _, member := range *members {
		memberEmails = append(memberEmails, member.Profile.Email)
	}

	return helpers.RemoveEmpty(memberEmails), err
}

func (o *Client) getGroupsMembers(groupID string) (*GroupMembers, *http.Response, error) {
	req, err := o.newRequest(http.MethodGet, fmt.Sprintf("%s/%s/users", groupBase, groupID), nil, nil)

	if err != nil {
		return nil, nil, err
	}

	var groupMembers GroupMembers
	resp, err := o.do(req, &groupMembers)
	if err != nil {
		return nil, nil, err
	}

	return &groupMembers, resp, nil
}
