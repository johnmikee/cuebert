package kandji

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/logger"
)

type Config struct {
	token   string
	baseURL string

	client *http.Client
	log    logger.Logger
}

// Setup implements mdm.Provider.
func (c *Config) Setup(m mdm.Config) {
	c.token = helpers.TokenValidator(m.Token, "Bearer")
	c.baseURL = helpers.URLShaper(m.URL, "api/v1/")
	c.client = httpClient(m.Client)
	c.log = logger.ChildLogger("kandji", &m.Log)
}

// GetDevice implements mdm.Provider.
func (c *Config) GetDevice(deviceID string) (*mdm.Device, error) {
	return c.GetDeviceDetails(deviceID)
}

// ListDevices implements mdm.Provider.
func (c *Config) ListDevices() ([]mdm.Device, error) {
	return c.ListAllDevices()
}

// Client stores the config needed to interact with the Kandji API.
type Client struct {
	Token   string
	BaseURL string

	Client *http.Client
	Log    *logger.Logger
}

type offsetRange struct {
	Limit  int
	Offset int
}

// NewClient returns a pointer with the Client after validating the arguments passedev[i].
func NewClient(c *Client) *Config {
	return &Config{
		token:   helpers.TokenValidator(c.Token, "Bearer"),
		baseURL: helpers.URLShaper(c.BaseURL, "api/v1/"),
		client:  httpClient(c.Client),
		log:     logger.ChildLogger("kandji", c.Log),
	}
}

func httpClient(c *http.Client) *http.Client {
	if c != nil {
		return c
	}

	return &http.Client{}
}

func (c *Config) newRequest(method, url string, body interface{}) (*http.Request, error) {
	var buf bytes.Buffer

	u := fmt.Sprintf("%s%s", c.baseURL, strings.TrimPrefix(url, "/"))

	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			c.log.Err(err).Msg("error encoding payload body")
			return nil, err
		}
	}
	req, err := http.NewRequest(method, u, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", c.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-type", "application/json;charset=utf-8")
	return req, nil
}

func (c *Config) do(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return resp, errors.Errorf("error during request. status code=%d error: %s", resp.StatusCode, string(body))
	}

	return resp, nil
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
			LastCheckIn:     dev[i].LastCheckIn,
			User:            setMDMUser(dev[i].User),
			AssetTag:        dev[i].AssetTag,
			FirstEnrollment: dev[i].FirstEnrollment,
			LastEnrollment:  dev[i].LastEnrollment,
		})
	}
	return results
}
