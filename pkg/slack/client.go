package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"golang.org/x/time/rate"
)

type SlackClient struct {
	token   string
	baseURL string

	client      *http.Client
	ratelimiter *rate.Limiter
	sc          *slack.Client
	log         logger.Logger
}

func NewClient(token, url string, c *http.Client, l *logger.Logger) *SlackClient {
	return &SlackClient{
		token:       helpers.TokenValidator(token, "Bearer"),
		baseURL:     helpers.URLShaper(url, "api/"),
		client:      httpClient(c),
		ratelimiter: rate.NewLimiter(rate.Every(10*time.Second), 50),
		sc:          &slack.Client{},
		log:         logger.ChildLogger("slack", l),
	}
}

func httpClient(c *http.Client) *http.Client {
	if c != nil {
		return c
	}

	return http.DefaultClient
}

func (s *SlackClient) newRequest(method, url string, body interface{}) (*http.Request, error) {
	var buf bytes.Buffer

	u := fmt.Sprintf("%s%s", s.baseURL, strings.TrimPrefix(url, "/"))

	if body != nil {
		fmt.Println("body", body)
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			fmt.Println("err", err)
			s.log.Err(err).Msg("error encoding payload body")
			return nil, err
		}
	}
	req, err := http.NewRequest(method, u, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", s.token)

	return req, nil
}

func (s *SlackClient) do(req *http.Request, into interface{}) error {
	ctx := context.Background()
	s.ratelimiter.Wait(ctx)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == 429 {
			return fmt.Errorf("rate-limited")
		}
		body, _ := io.ReadAll(resp.Body)
		return errors.Errorf("unexpected response. status=%d api error: %s", resp.StatusCode, string(body))
	}

	err = json.NewDecoder(resp.Body).Decode(into)
	return errors.Wrap(err, "decoding Slack response response body")
}
