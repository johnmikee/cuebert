package mdm

import "time"

// Device holds the general purpose information of the device
type Device struct {
	DeviceID        string      `json:"device_id"`
	DeviceName      string      `json:"device_name"`
	Model           string      `json:"model"`
	SerialNumber    string      `json:"serial_number"`
	Platform        string      `json:"platform"`
	OSVersion       string      `json:"os_version"`
	LastCheckIn     *time.Time  `json:"last_check_in"`
	User            User        `json:"user"`
	AssetTag        interface{} `json:"asset_tag"`
	FirstEnrollment string      `json:"first_enrollment"`
	LastEnrollment  string      `json:"last_enrollment"`
}

type DeviceResults []Device

type DeviceDetails struct {
	General Device `json:"general"`
	Users   User   `json:"users"`
}

type User struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	ID    string `json:"id"`
}

// QueryOpts is a list of options to query devices
type QueryOpts struct {
	AssetTag     string `json:"asset_tag"`
	DeviceID     string `json:"device_id"`
	DeviceName   string `json:"device_name"`
	Model        string `json:"model"`
	OSVersion    string `json:"os_version"`
	SerialNumber string `json:"serial_number"`
	Platform     string `json:"platform"`
	UserEmail    string `json:"user_email"`
	UserID       string `json:"user_id"`
	UserName     string `json:"user_name"`
}
