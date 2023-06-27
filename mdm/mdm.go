package mdm

import (
	"net/http"

	"github.com/johnmikee/cuebert/pkg/logger"
)

// MDM represents the type of MDM provider.
type MDM string

const (
	Jamf   MDM = "jamf"
	Kandji MDM = "kandji"
)

// Provider represents the interface for an MDM provider.
type Provider interface {
	Setup(config Config)
	ListDevices() ([]Device, error)
	GetDevice(deviceID string) (*Device, error)
	QueryDevices(opts *QueryOpts) (DeviceResults, error)
	GetUsers(opts *QueryOpts) ([]User, error)
}

type Config struct {
	// Common configuration fields
	Domain   string        `json:"domain,omitempty"`
	MDM      MDM           `json:"mdm,omitempty"`
	URL      string        `json:"url,omitempty"`
	User     string        `json:"user,omitempty"`
	Password string        `json:"password,omitempty"`
	Token    string        `json:"token,omitempty"`
	Client   *http.Client  `json:"client,omitempty"`
	Log      logger.Logger `json:"log,omitempty"`
	// Provider-specific configuration fields
	ProviderSpecificConfig interface{} `json:"provider_specific_config,omitempty"`
}
