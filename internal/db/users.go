package db

import (
	"github.com/johnmikee/cuebert/db/users"
)

type User = Config

// AddUser provides access to the UserUpdate.Add() method
func (u *User) AddUser() *users.UserUpdate {
	return u.user(u.db, &u.log).Add()
}

// UserBy provides access to the UserQuery.By() method
func (u *User) UserBy() *users.UserQuery {
	return u.user(u.db, &u.log).By()
}

// RemoveUserBy provides access to UserRemove
func (u *User) RemoveUserBy() *users.UserRemove {
	return u.user(u.db, &u.log).Remove()
}

// AddAllUsers adds all users to the users table
func (u *User) AddAllUsers(us users.UI) error {
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
