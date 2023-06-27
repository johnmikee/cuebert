package client

import (
	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/mdm/jamf"
	"github.com/johnmikee/cuebert/mdm/kandji"
)

// Config represents the configuration for the client.
type Config struct {
	MDMProvider mdm.Provider
}

// MDM represents the MDM client.
type MDM struct {
	MDM    mdm.MDM
	Config mdm.Config
}

// New creates a new MDM provider based on the provided MDM configuration.
// It returns the MDM provider instance.
func New(m *MDM) mdm.Provider {
	config := Config{
		MDMProvider: createMDMProvider(m.MDM),
	}

	config.MDMProvider.Setup(m.Config)

	return config.MDMProvider
}

// createMDMProvider creates and returns an MDM provider based on the provided MDM type.
func createMDMProvider(providerName mdm.MDM) mdm.Provider {
	switch providerName {
	case mdm.Jamf:
		return &jamf.Client{}
	case mdm.Kandji:
		return &kandji.Config{}
	default:
		return nil
	}
}
