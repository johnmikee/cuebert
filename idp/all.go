package idp

import "time"

// User holds information on the idp user
type User struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Activated time.Time `json:"activated"`
	Profile   Profile   `json:"profile"`
}

// Profiles holds information on the users profile
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
}
