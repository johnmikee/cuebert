package client

import (
	"github.com/johnmikee/cuebert/idp"
	"github.com/johnmikee/cuebert/idp/okta"
)

type Config struct {
	IDPProvider idp.Provider
}

type IDP struct {
	IDP    idp.IDP
	Config idp.Config
}

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
