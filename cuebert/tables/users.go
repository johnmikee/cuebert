package tables

import (
	"github.com/johnmikee/cuebert/db/users"
)

type User = Config

// AddUser provides access to the Update.Add() method
func (u *User) AddUser() *users.Update {
	return u.user(u.db, &u.log).Add()
}

// UserBy provides access to the Query.By() method
func (u *User) UserBy() *users.Query {
	return u.user(u.db, &u.log).By()
}

// RemoveUserBy provides access to Remove
func (u *User) RemoveUserBy() *users.Remove {
	return u.user(u.db, &u.log).Remove()
}

// AddAllUsers adds all users to the users table
func (u *User) UsersAddAll(us users.UI) error {
	_, err := u.user(u.db, &u.log).AddAllUsers(us)

	return err
}

// GetAllUsers returns all users from the users table
func (u *User) GetAllUsers() (users.UI, error) {
	return u.user(u.db, &u.log).By().All().Query()
}

// UserByEmail returns a user by email from the users table
func (u *User) UserByEmail(user string) (users.UI, error) {
	return u.user(u.db, &u.log).By().Email(user).Query()
}

// UserByID returns a user by ID from the users table
func (u *User) UserByID(user string) (users.UI, error) {
	return u.user(u.db, &u.log).By().SlackID(user).Query()
}

// UpdateUserBy returns
func (u *User) UpdateUserBy() *users.Update {
	return u.user(u.db, &u.log).Update()
}
