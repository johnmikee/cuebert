package okta

import (
	"fmt"
	"net/http"
	"time"
)

// Constants for the base endpoint of device assurance policies.
const deviceAssuranceBase = "device-assurances"

// DeviceAssurancePolicies represents a slice of device assurance policy structures.
type DeviceAssurancePolicies []struct {
	CreatedBy       string `json:"createdBy"`
	CreatedDate     string `json:"createdDate"`
	ID              string `json:"id"`
	LastUpdatedBy   string `json:"lastUpdatedBy"`
	LastUpdatedDate string `json:"lastUpdatedDate"`
	Name            string `json:"name"`
	Platform        string `json:"platform"`
	Links           Links  `json:"_links"`
}

// DAComputerPolicy represents a device assurance computer policy structure.
type DAComputerPolicy struct {
	ID                    string             `json:"id"`
	Name                  string             `json:"name"`
	LastUpdate            time.Time          `json:"lastUpdate"`
	CreatedUpdate         time.Time          `json:"createdUpdate"`
	LastUpdatedBy         string             `json:"lastUpdatedBy"`
	CreatedBy             string             `json:"createdBy"`
	OsVersion             OsVersion          `json:"osVersion"`
	DiskEncryptionType    DiskEncryptionType `json:"diskEncryptionType"`
	Jailbreak             bool               `json:"jailbreak"`
	Platform              string             `json:"platform"`
	ScreenLockType        ScreenLockType     `json:"screenLockType"`
	SecureHardwarePresent bool               `json:"secureHardwarePresent"`
	Links                 Links              `json:"_links"`
}

// OsVersion represents the minimum required operating system version.
type OsVersion struct {
	Minimum string `json:"minimum"`
}

// DiskEncryptionType represents disk encryption types for a policy.
type DiskEncryptionType struct {
	Include []string `json:"include"`
}

// ScreenLockType represents screen lock types for a policy.
type ScreenLockType struct {
	Include []string `json:"include"`
}

// Platform represents device platforms.
type Platform string

// Available platform constants.
const (
	Android Platform = "ANDROID"
	IOS     Platform = "IOS"
	Windows Platform = "WINDOWS"
	MacOS   Platform = "MACOS"
)

// AddToBlocked adds a user to the blocked group.
func (c *Client) AddToBlocked(userID string) error {
	_, err := c.addUserToGroup(c.blockGroup, userID)

	return err
}

// RemoveFromBlocked removes a user from the blocked group.
func (c *Client) RemoveFromBlocked(user string) error {
	c.log.Info().Str("user_id", user).Msg("getting user id")
	u, err := c.user(user)
	if err != nil {
		return fmt.Errorf("getting user id: %w", err)
	}
	_, err = c.removeUserFromGroup(c.blockGroup, u.ID)
	c.log.Info().AnErr("error", err).Msg("removing user from blocked group")
	return err
}

// DeviceAssurancePolicies returns a reference to the DeviceAssurancePolicy struct.
func (o *Client) DeviceAssurancePolicies() *DeviceAssurancePolicy {
	return &DeviceAssurancePolicy{
		client: o,
	}
}

// DeviceAssurancePolicy represents a client for managing device assurance policies.
type DeviceAssurancePolicy struct {
	client *Client
}

// Create creates a new device assurance computer policy.
func (d *DeviceAssurancePolicy) Create() (*DAComputerPolicy, error) {
	req, err := d.client.newRequest(http.MethodPost, deviceAssuranceBase, false, nil)
	if err != nil {
		return nil, err
	}

	var dap DAComputerPolicy
	_, err = d.client.do(req, &dap)
	if err != nil {
		return nil, err
	}

	return &dap, nil
}

// List retrieves a list of device assurance policies.
func (d *DeviceAssurancePolicy) List() (DeviceAssurancePolicies, error) {
	req, err := d.client.newRequest(http.MethodGet, deviceAssuranceBase, false, nil)
	if err != nil {
		return nil, err
	}

	var dap DeviceAssurancePolicies
	_, err = d.client.do(req, &dap)
	if err != nil {
		return nil, err
	}

	return dap, nil
}

// Get retrieves a specific device assurance computer policy by ID.
func (d *DeviceAssurancePolicy) Get(id string) (*DAComputerPolicy, error) {
	req, err := d.client.newRequest(http.MethodGet, deviceAssuranceBase+"/"+id, false, nil)
	if err != nil {
		return nil, err
	}

	var dap DAComputerPolicy
	_, err = d.client.do(req, &dap)
	if err != nil {
		return nil, err
	}

	return &dap, nil
}

// Update updates a device assurance computer policy.
func (d *DeviceAssurancePolicy) Update(policy DAComputerPolicy) (*DAComputerPolicy, error) {
	req, err := d.client.newRequest(http.MethodPut, deviceAssuranceBase+"/"+policy.ID, false, policy)
	if err != nil {
		return nil, err
	}

	var dap DAComputerPolicy
	_, err = d.client.do(req, &dap)
	if err != nil {
		return nil, err
	}

	return &dap, nil
}
