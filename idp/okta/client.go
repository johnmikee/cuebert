package okta

import (
	"net/http"

	"github.com/johnmikee/cuebert/idp"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/johnmikee/cuebert/pkg/requester"
)

// Client represents the Okta client.
type Client struct {
	token      string
	baseURL    string
	domain     string
	blockGroup string
	client     *http.Client
	log        logger.Logger
}

// Config represents the configuration for the Okta client.
type Config struct {
	BlockGroup string         `json:"block_group,omitempty"`
	Domain     string         `json:"domain,omitempty"`
	URL        string         `json:"url,omitempty"`
	Token      string         `json:"token,omitempty"`
	Client     *http.Client   `json:"client,omitempty"`
	Log        *logger.Logger `json:"log,omitempty"`
}

// Setup implements idp.Provider configuring the okta client.
func (c *Client) Setup(i idp.Config) {
	c.token = helpers.TokenValidator(i.Token, "SSWS")
	c.baseURL = helpers.URLShaper(i.URL, "api/v1/")
	c.domain = i.Domain
	c.client = i.Client
	c.log = logger.ChildLogger("idp/okta", &i.Log)
	c.blockGroup = i.IDPBlockGroup
}

// New returns a pointer with the Client after validating the arguments passed.
func New(c *Config) *Client {
	return &Client{
		token:      helpers.TokenValidator(c.Token, "SSWS"),
		baseURL:    helpers.URLShaper(c.URL, "api/v1/"),
		domain:     c.Domain,
		client:     c.Client,
		log:        logger.ChildLogger("idp/okta", c.Log),
		blockGroup: c.BlockGroup,
	}
}

func (o *Client) newRequest(method, url string, override bool, body interface{}) (*http.Request, error) {
	return requester.New(method, o.baseURL, url, override, body)
}

func (o *Client) headers(req *http.Request) *http.Request {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", o.token)

	return req
}

func (o *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	o.headers(req)
	return requester.Do(o.client, req, v)
}
