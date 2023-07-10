package tables

import (
	"errors"
	"time"

	"github.com/johnmikee/cuebert/db/exclusions"
	"github.com/johnmikee/cuebert/pkg/helpers"
)

type Exclusion = Config

func (e *Exclusion) AddExclusions() *exclusions.Update {
	return e.exclusions(e.db, &e.log).Add()
}

// AddExclusion adds an Exclusion to the database
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

// ApproveExclusion approves an Exclusion
func (e *Exclusion) ApproveExclusion(reason, serial string, until time.Time) error {
	_, err := e.exclusions(e.db, &e.log).Update().
		Approved(true).
		Reason(reason).
		SerialNumber(serial).
		Until(until).
		Parse("serial_number", serial).
		Send()

	return err
}

func (e *Exclusion) ExclusionSerials(slackID string) []string {
	resp, err := e.GetUsersSerialsBot(slackID)

	if err != nil {
		e.log.Error().Err(err).Msg("could not get serials")
		return nil
	}

	serials := []string{}
	for i := range resp {
		serials = append(serials, resp[i].SerialNumber)
	}

	return serials
}

// RemoveExclusion removes an Exclusion from the database
func (e *Exclusion) RemoveExclusion() *exclusions.Remove {
	return e.exclusions(e.db, &e.log).Remove()
}

// ExlusionBy provides access to querying the Exclusions table
func (e *Exclusion) ExclusionBy() *exclusions.Query {
	return e.exclusions(e.db, &e.log).Query()
}

// RequestExclusion allows a user to request an Exclusion
func (e *Exclusion) RequestExclusion(sid, reason string, serials []string, until time.Time) error {
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

// SerialExcluded returns any Exclusions for a given serial number
func (e *Exclusion) SerialExcluded(serial string) (exclusions.EI, error) {
	return e.exclusions(e.db, &e.log).Query().Serial(serial).Query()
}

// SetExclusion sets an Exclusion for a given slackid
func (e *Exclusion) SetExclusion(sid, reason string, until time.Time) error {
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
