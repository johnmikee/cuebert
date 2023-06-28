package jamf

import (
	"net/http"
	"time"

	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/logger"
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

type Client struct {
	domain   string
	username string
	password string
	url      string
	log      logger.Logger
	client   *http.Client
}

// GetUsers implements mdm.Provider.
func (*Client) GetUsers(opts *mdm.QueryOpts) ([]mdm.User, error) {
	panic("unimplemented")
}

// QueryDevices implements mdm.Provider.
func (*Client) QueryDevices(opts *mdm.QueryOpts) (mdm.DeviceResults, error) {
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

// Setup implements mdm.Provider.
func (c *Client) Setup(m mdm.Config) {
	c.domain = m.Domain
	c.username = m.User
	c.password = m.Password
	c.url = helpers.URLShaper(m.URL, "JSSResource")
	c.log = logger.ChildLogger("jamf", &m.Log)
	c.client = httpClient(m.Client)
}

func httpClient(c *http.Client) *http.Client {
	if c != nil {
		return c
	}

	return defaultHTTPClient()
}

// Used if custom client not passed on when NewClient instantiated
func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: time.Minute,
	}
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
