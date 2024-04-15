package info

import "github.com/johnmikee/cuebert/cuebert/task/implementation"

// Users stores the basic information for users in each task used to populate the DB
type Users struct {
	ID      int    `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	SlackID string `json:"slack_id"`
}

// Tasks store information on the given task at hand
type Tasks struct {
	UserTaskID      int    `json:"user_task_id"`
	UserID          int    `json:"user_id"`
	UserEmail       string `json:"user_email"`
	TaskID          int    `json:"task_id"`
	ItemName        string `json:"item_name"`
	TaskDescription string `json:"task_description"`
	TaskComplete    bool   `json:"task_complete"`
}

// UL is a custom type for a slice of Users for convenience
type UL []Users

// TL is a custom type for a slice of Tasks for convenience
type TL []Tasks

// PL is a custom type for a slice of PersistUsers for convenience
type PL []PersistUsers

// ToUsers converts a slice of Tasks to a slice of Users
func (t TL) ToUsers() []Users {
	users := []Users{}
	for _, u := range t {
		users = append(users, Users{
			ID:    u.UserID,
			Email: u.UserEmail,
		})
	}
	return users
}

// Persist converts a slice of Tasks to a slice of PersistUsers
func (t TL) Persist(i implementation.Item) []PersistUsers {
	users := []PersistUsers{}
	for _, u := range t {
		users = append(users, PersistUsers{
			UserID:    u.UserID,
			UserEmail: u.UserEmail,
			Task:      i,
		})
	}
	return users
}

// PersistUsers is used to track users across tasks who should _still_ be present
type PersistUsers struct {
	UserID    int                 `json:"user_id"`
	UserEmail string              `json:"user_email"`
	Task      implementation.Item `json:"task"`
}

// UsersSlice returns a slice of user ids for the given users
func (p PL) UsersSlice() []int {
	users := []int{}
	for _, u := range p {
		users = append(users, u.UserID)
	}
	return users
}

// EmailSlice returns a slice of emails for the given users
func (p PL) EmailSlice() []string {
	users := []string{}
	for _, u := range p {
		users = append(users, u.UserEmail)
	}
	return users
}

// EmailMap returns a map of emails for the given users
func (p PL) EmailMap() map[string]struct{} {
	emailMap := make(map[string]struct{})
	for _, email := range p {
		emailMap[email.UserEmail] = struct{}{}
	}
	return emailMap
}
