package google

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/johnmikee/cuebert/pkg/logger"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Config holds the configuration needed to set up the GoogleClient
type Config struct {
	GoogleServiceCreds string
	Scopes             []string
	AdminEmail         string
	CustomerID         string
	SheetID            string
	Log                logger.Logger
}

// Client holds the config for setting up our client
// to interact with the Google Sheets API
type Client struct {
	ctx     context.Context
	ts      oauth2.TokenSource
	sheets  *sheets.Service
	sheetID string
	config  *Config
	log     logger.Logger
}

func New(c *Config, opts ...Option) (*Client, error) {
	gc, err := buildClient(c)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(gc)
	}

	return gc, nil
}

type Option func(c *Client)

func WithSheets() Option {
	return func(g *Client) {
		g.newSheets()
	}
}

func (c *Client) newSheets() {
	sheetsService, err := sheets.NewService(c.ctx, option.WithTokenSource(c.ts))
	if err != nil {
		c.log.Debug().AnErr("error", err).Msg("Error generating cloud client")
	}

	c.sheets = sheetsService
}

// credsToContent takes the base64 encoded string of the google service file
// and returns a string version of the contents
func credsToContent(c string) ([]byte, error) {
	content, err := base64.StdEncoding.DecodeString(c)
	if err != nil {
		return nil, fmt.Errorf("error decoding base64 string %s", err.Error())
	}

	return content, nil
}

func (c *Config) genClient(ctx context.Context) (oauth2.TokenSource, error) {
	creds, err := credsToContent(c.GoogleServiceCreds)
	if err != nil {
		return nil, fmt.Errorf("error decoding contents of google service creds %s", err.Error())
	}

	config, err := google.JWTConfigFromJSON(creds, c.Scopes...)
	if err != nil {
		return nil, fmt.Errorf("error generating JWT config from JSON %s", err.Error())
	}

	config.Subject = c.AdminEmail
	ts := config.TokenSource(ctx)

	return ts, nil
}

func buildClient(c *Config) (*Client, error) {
	ctx := context.Background()
	cl, err := c.genClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating google client: %s", err.Error())
	}
	gc := &Client{
		ctx:     ctx,
		ts:      cl,
		config:  c,
		sheetID: c.SheetID,
		log:     logger.ChildLogger("sheets", &c.Log),
	}

	return gc, nil
}
