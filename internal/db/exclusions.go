package db

import (
	"errors"
	"time"

	"github.com/johnmikee/cuebert/db/exclusions"
	"github.com/johnmikee/cuebert/pkg/helpers"
)

type Exclusion = Config

func (e *Exclusion) AddExclusions() *exclusions.ExclusionUpdate {
	return e.exclusions(e.db, &e.log).Add()
}

// AddExclusion adds an exclusion to the database
func (e *Exclusion) AddExclusion(serial, reason string, until time.Time) error {
	di, err := e.DeviceBySerial(serial)
	if err != nil {
		return err
	}

	if di.Empty() {
		return errors.New("no device found with that serial number")
	}

	email := di[0].User

	_, err = e.exclusions(e.db, &e.log).Add().
		Approved(true).
		Email(email).
		Reason(reason).
		SerialNumber(serial).
		Until(until).
		Execute()

	return err
}

// ApproveException approves an exception
func (e *Exclusion) ApproveException(reason, serial string, until time.Time) error {
	_, err := e.exclusions(e.db, &e.log).Update().
		Approved(true).
		Reason(reason).
		SerialNumber(serial).
		Until(until).
		Parse("serial_number", serial).
		Send()

	return err
}

// RemoveExclusion removes an exclusion from the database
func (e *Exclusion) RemoveExclusion() *exclusions.ExclusionRemove {
	return e.exclusions(e.db, &e.log).Remove()
}

// ExlusionBy provides access to querying the exclusions table
func (e *Exclusion) ExclusionBy() *exclusions.ExclusionQuery {
	return e.exclusions(e.db, &e.log).By()
}

// RequestException allows a user to request an exception
func (e *Exclusion) RequestException(sid, reason string, serials []string, until time.Time) error {
	user, err := e.br(e.db, &e.log).Query().SlackID(sid).Query()
	if err != nil {
		return err
	}

	if user.Empty() {
		return errors.New("no record found with that slack id")
	}

	for i := range user {
		if !helpers.Contains(serials, user[i].SerialNumber) {
			continue
		}

		_, err = e.exclusions(e.db, &e.log).Add().
			Approved(false).
			Email(user[i].UserEmail).
			Reason(reason).
			SerialNumber(user[i].SerialNumber).
			Until(until).
			Execute()
	}

	return err
}

// SerialExcluded returns any exclusions for a given serial number
func (e *Exclusion) SerialExcluded(serial string) ([]exclusions.ExclusionInfo, error) {
	return e.exclusions(e.db, &e.log).By().Serial(serial).Query()
}

// SetException sets an exception for a given slackid
func (e *Exclusion) SetException(sid, reason string, until time.Time) error {
	user, err := e.DeviceBySerial(sid)
	if err != nil {
		return err
	}

	if user.Empty() {
		return errors.New("no device found with that serial number")
	}

	email := user[0].User
	serial := user[0].SerialNumber

	_, err = e.exclusions(e.db, &e.log).Add().
		Email(email).
		Reason(reason).
		SerialNumber(serial).
		Until(until).
		Execute()

	return err
}
