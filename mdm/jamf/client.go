package jamf

import (
	"net/http"

	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/johnmikee/cuebert/pkg/requester"
)

// Client represents the configuration for the Jamf API client
type Config struct {
	Domain   string
	Username string
	Password string
	BaseURL  string
	Log      logger.Logger
	client   *http.Client
}

// Client represents the Jamf API client
type Client struct {
	domain   string
	username string
	password string
	url      string
	log      logger.Logger
	client   *http.Client
}

// Setup implements mdm.Provider.
func (c *Client) Setup(m mdm.Config) {
	c.domain = m.Domain
	c.username = m.User
	c.password = m.Password
	c.url = helpers.URLShaper(m.URL, "JSSResource")
	c.log = logger.ChildLogger("jamf", &m.Log)
	c.client = m.Client
}

// New returns a client to communicate with the Jamf API
func New(c *Config) *Client {
	return &Client{
		domain:   c.Domain,
		username: c.Username,
		password: c.Password,
		url:      c.BaseURL,
		log:      c.Log,
		client:   c.client,
	}
}

// GetUsers implements mdm.Provider.
func (c *Client) GetUsers(opts *mdm.QueryOpts) ([]mdm.User, error) {
	panic("unimplemented")
}

// QueryDevices implements mdm.Provider.
func (c *Client) QueryDevices(opts *mdm.QueryOpts) (mdm.DeviceResults, error) {
	panic("unimplemented")
}

// GetDevice implements mdm.Provider.
func (c *Client) GetDevice(deviceID string) (*mdm.Device, error) {
	panic("unimplemented")
}

// ListDevices implements mdm.Provider.
func (c *Client) ListDevices() ([]mdm.Device, error) {
	panic("unimplemented")
}

func (c *Client) newRequest(method, url string, override bool, body interface{}) (*http.Request, error) {
	return requester.New(method, c.url, url, override, body)
}

func (c *Client) headers(req *http.Request) *http.Request {
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Accept", "application/json")

	return req
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	c.headers(req)
	return requester.Do(c.client, req, v)
}
