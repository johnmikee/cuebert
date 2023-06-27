package idp

import (
	"net/http"

	"github.com/johnmikee/cuebert/pkg/logger"
)

type IDP string

const (
	Okta IDP = "okta"
)

type Provider interface {
	Setup(config Config)
	GetAdminGroup(string) ([]string, error)
	GetAllUsers() ([]User, error)
}

type Config struct {
	Domain string        `json:"domain,omitempty"`
	URL    string        `json:"url,omitempty"`
	Token  string        `json:"token,omitempty"`
	Client *http.Client  `json:"client,omitempty"`
	Log    logger.Logger `json:"log,omitempty"`
}
