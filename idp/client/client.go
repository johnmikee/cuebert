package client

import (
	"github.com/johnmikee/cuebert/idp"
	"github.com/johnmikee/cuebert/idp/okta"
)

// Config represents the configuration for the client.
type Config struct {
	IDPProvider idp.Provider
}

// IDP represents the IDP client storing the type and configuration.
type IDP struct {
	IDP    idp.IDP
	Config idp.Config
}

// New returns a new IDP provider based on the provided IDP configuration returning the IDP provider instance.
func New(i *IDP) idp.Provider {
	config := Config{
		IDPProvider: createIDPProvider(i.IDP),
	}

	config.IDPProvider.Setup(i.Config)

	return config.IDPProvider
}

func createIDPProvider(providerName idp.IDP) idp.Provider {
	switch providerName {
	case idp.Okta:
		return &okta.Client{}
	default:
		return nil
	}
}
