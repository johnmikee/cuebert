package idp

import (
	"net/http"

	"github.com/johnmikee/cuebert/pkg/logger"
)

// IDP represents the IDP type.
type IDP string

const (
	Okta IDP = "okta"
)

// Provider represents the IDP provider.
type Provider interface {
	Setup(config Config)
	GetAdminGroup(string) ([]string, error)
	GetAllUsers() (UI, error)
	GetUser(string) (*User, error)
	AddToBlocked(string) error
	RemoveFromBlocked(string) error
}

// Config is used to configure the IDP provider.
type Config struct {
	IDPBlockGroup string        `json:"block_group,omitempty"`
	Domain        string        `json:"domain,omitempty"`
	URL           string        `json:"url,omitempty"`
	Token         string        `json:"token,omitempty"`
	Client        *http.Client  `json:"client,omitempty"`
	Log           logger.Logger `json:"log,omitempty"`
}
