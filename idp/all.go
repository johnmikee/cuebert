package idp

import (
	"time"
)

// User holds information on the idp user.
type User struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Activated time.Time `json:"activated"`
	Profile   Profile   `json:"profile"`
}

// Profiles holds information on the users profile.
type Profile struct {
	LastName   string `json:"lastName"`
	Manager    string `json:"manager"`
	ManagerID  string `json:"managerId"`
	Title      string `json:"title"`
	Login      string `json:"login"`
	FirstName  string `json:"firstName"`
	UserType   string `json:"userType"`
	Department string `json:"department"`
	Email      string `json:"email"`
	StartDate  string `json:"startDate"`
}

// UI is a slice of User used for convenience functions.
type UI []User

// MakeEmailProfileMap returns a map of email to profile from a slice of users.
func (u UI) MakeEmailProfileMap() map[string]Profile {
	epm := make(map[string]Profile)
	for i := range u {
		epm[u[i].Profile.Email] = u[i].Profile
	}
	return epm
}

// MakeEmailIDMap returns a map of email to id from a slice of users.
func (u UI) MakeEmailIDMap() map[string]string {
	eim := make(map[string]string)
	for i := range u {
		eim[u[i].Profile.Email] = u[i].ID
	}
	return eim
}

// MakeEmailManagerMap returns a map of email to manager email.
func (u UI) MakeEmailManagerMap() map[string]string {
	emm := make(map[string]string)
	for i := range u {
		emm[u[i].Profile.Email] = u[i].Profile.ManagerID
	}
	return emm
}
