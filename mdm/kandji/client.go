package kandji

import (
	"net/http"
	"strconv"

	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/johnmikee/cuebert/pkg/requester"
)

// Config stores the config needed to interact with the Kandji API.
type Config struct {
	Token   string
	BaseURL string
	Client  *http.Client
	Log     *logger.Logger
}

// Client represents the Kandji API client
type Client struct {
	token   string
	baseURL string
	client  *http.Client
	log     logger.Logger
}

// Setup implements mdm.Provider.
func (c *Client) Setup(m mdm.Config) {
	c.client = m.Client
	c.log = logger.ChildLogger("kandji", &m.Log)
	c.baseURL = helpers.URLShaper(m.URL, "api/v1/")
	c.token = helpers.TokenValidator(m.Token, "Bearer")
}

// GetDevice returns a device from the Kandji API.
//
//	func (c *Client) GetDevice(deviceID string) (*mdm.Device, error) {
//		return c.GetDeviceDetails(deviceID)
//	}
func (c *Client) GetDevice(deviceID string) (*mdm.Device, error) {
	dg, err := c.getDevice(deviceID)
	if err != nil {
		return nil, err
	}

	return &mdm.Device{
		DeviceID:        dg.DeviceID,
		DeviceName:      dg.DeviceName,
		Model:           dg.Model,
		SerialNumber:    dg.SerialNumber,
		Platform:        dg.Platform,
		OSVersion:       dg.OsVersion,
		LastCheckIn:     &dg.LastCheckIn,
		AssetTag:        dg.AssetTag,
		FirstEnrollment: dg.FirstEnrollment,
		LastEnrollment:  dg.LastEnrollment,
		User: mdm.User{
			Email: dg.User.Email,
			Name:  dg.User.Name,
			ID:    strconv.Itoa(dg.User.ID),
		},
	}, nil
}

// ListDevices returns a slice of devices formatted for the MDM interface.
func (c *Client) ListDevices() ([]mdm.Device, error) {
	return c.ListAllDevices()
}

type offsetRange struct {
	Limit  int
	Offset int
}

// NewClient returns a pointer with the Client after validating the arguments passedev[i].
func NewClient(c *Config) *Client {
	return &Client{
		token:   helpers.TokenValidator(c.Token, "Bearer"),
		baseURL: helpers.URLShaper(c.BaseURL, "api/v1/"),
		client:  c.Client,
		log:     logger.ChildLogger("kandji", c.Log),
	}
}

func (c *Client) newRequest(url string) (*http.Request, error) {
	return requester.New("GET", c.baseURL, url, false, nil)
}

func (c *Client) headers(req *http.Request) {
	req.Header.Set("Authorization", c.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-type", "application/json;charset=utf-8")
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	c.headers(req)
	return requester.Do(c.client, req, v)
}

func transform(dev DeviceResults) mdm.DeviceResults {
	var results mdm.DeviceResults

	for i := range dev {
		results = append(results, mdm.Device{
			DeviceID:        dev[i].DeviceID,
			DeviceName:      dev[i].DeviceName,
			Model:           dev[i].Model,
			SerialNumber:    dev[i].SerialNumber,
			Platform:        dev[i].Platform,
			OSVersion:       dev[i].OSVersion,
			LastCheckIn:     kandjiTimeFixer(dev[i].LastCheckIn),
			User:            setMDMUser(dev[i].User),
			AssetTag:        dev[i].AssetTag,
			FirstEnrollment: dev[i].FirstEnrollment,
			LastEnrollment:  dev[i].LastEnrollment,
		})
	}
	return results
}
