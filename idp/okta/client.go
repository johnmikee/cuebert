package okta

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/johnmikee/cuebert/idp"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// Client represents the Okta client.
type Client struct {
	token   string
	baseURL string
	domain  string

	client *http.Client
	log    logger.Logger
}

// Config represents the configuration for the Okta client.
type Config struct {
	Domain string         `json:"domain,omitempty"`
	URL    string         `json:"url,omitempty"`
	Token  string         `json:"token,omitempty"`
	Client *http.Client   `json:"client,omitempty"`
	Log    *logger.Logger `json:"log,omitempty"`
}

// GetAllUsers implements idp.Provider.
func (c *Client) GetAllUsers() ([]idp.User, error) {
	return c.getAllUsers()
}

// Setup implements idp.Provider.
func (c *Client) Setup(i idp.Config) {
	c.token = helpers.TokenValidator(i.Token, "SSWS")
	c.baseURL = helpers.URLShaper(i.URL, "api/v1/")
	c.domain = i.Domain
	c.client = httpClient(i.Client)
	c.log = logger.ChildLogger("idp/okta", &i.Log)
}

// urlOverride is used to override the default url arg in a function
type urlOverride struct {
	override bool
	url      string
}

// NewClient returns a pointer with the Client after validating the arguments passed.
func NewClient(c *Config) *Client {
	return &Client{
		token:   helpers.TokenValidator(c.Token, "SSWS"),
		baseURL: helpers.URLShaper(c.URL, "api/v1/"),
		domain:  c.Domain,
		client:  httpClient(c.Client),
		log:     logger.ChildLogger("idp/okta", c.Log),
	}
}

func httpClient(c *http.Client) *http.Client {
	if c != nil {
		return c
	}

	return &http.Client{}
}

func (o *Client) setHeaders(req *http.Request) *http.Request {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", o.token)

	return req
}

func (o *Client) newRequest(method, u string, ur *urlOverride, body interface{}) (*http.Request, error) {
	var buf bytes.Buffer

	url := fmt.Sprintf("%s%s", o.baseURL, strings.TrimPrefix(u, "/"))

	if ur != nil && ur.override {
		url = ur.url
	}
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			o.log.Debug().Err(err).Msg("error building body")
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		o.log.Err(err).Msg("error building request")
		return nil, err
	}

	req = o.setHeaders(req)

	return req, nil
}

func (o *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}

	// If StatusCode is not in the 200 range something went wrong, return the
	// response but do not process it's body.
	if c := resp.StatusCode; 200 > c || c > 299 {
		return resp, nil
	}

	defer resp.Body.Close()
	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err := io.Copy(w, resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to copy response body: %s", err.Error())
			}
		} else {
			decErr := json.NewDecoder(resp.Body).Decode(v)
			if decErr == io.EOF {
				decErr = nil // ignore EOF errors caused by empty response body
			}
			if decErr != nil {
				err = decErr
			}
		}
	}

	return resp, err
}
