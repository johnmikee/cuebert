package slack

import (
	"fmt"
	"net/http"
)

type User struct {
	ID        string  `json:"id"`
	TeamID    string  `json:"team_id"`
	Name      string  `json:"name"`
	Profile   Profile `json:"profile"`
	IsBot     bool    `json:"is_bot"`
	IsAppUser bool    `json:"is_app_user"`
}

type UserPagination struct {
	users []Member

	limit  int
	cursor string

	c *SlackClient
}

type SlackUserInfo struct {
	Ok   bool `json:"ok"`
	User User `json:"user"`
}

type Profile struct {
	RealName              string `json:"real_name"`
	RealNameNormalized    string `json:"real_name_normalized"`
	DisplayName           string `json:"display_name"`
	DisplayNameNormalized string `json:"display_name_normalized"`
	Email                 string `json:"email"`
	FirstName             string `json:"first_name"`
	LastName              string `json:"last_name"`
}

type ListResponse struct {
	Ok               bool             `json:"ok"`
	Members          []Member         `json:"members"`
	CacheTs          int64            `json:"cache_ts"`
	ResponseMetadata ResponseMetadata `json:"response_metadata"`
}

type ResponseMetadata struct {
	NextCursor string `json:"next_cursor"`
}

type Member struct {
	ID                string  `json:"id"`
	TeamID            string  `json:"team_id"`
	Name              string  `json:"name"`
	Deleted           bool    `json:"deleted"`
	Color             string  `json:"color"`
	RealName          string  `json:"real_name"`
	Tz                *string `json:"tz"`
	TzLabel           string  `json:"tz_label"`
	TzOffset          int64   `json:"tz_offset"`
	Profile           Profile `json:"profile"`
	IsAdmin           bool    `json:"is_admin"`
	IsOwner           bool    `json:"is_owner"`
	IsPrimaryOwner    bool    `json:"is_primary_owner"`
	IsRestricted      bool    `json:"is_restricted"`
	IsUltraRestricted bool    `json:"is_ultra_restricted"`
	IsBot             bool    `json:"is_bot"`
	Updated           int64   `json:"updated"`
	Has2Fa            *bool   `json:"has_2fa,omitempty"`
}

// UserProfile contains all the information details of a given user
type UserProfile struct {
	FirstName             string `json:"first_name"`
	LastName              string `json:"last_name"`
	RealName              string `json:"real_name"`
	RealNameNormalized    string `json:"real_name_normalized"`
	DisplayName           string `json:"display_name"`
	DisplayNameNormalized string `json:"display_name_normalized"`
	Email                 string `json:"email"`
	Skype                 string `json:"skype"`
	Phone                 string `json:"phone"`
	Image24               string `json:"image_24"`
	Image32               string `json:"image_32"`
	Image48               string `json:"image_48"`
	Image72               string `json:"image_72"`
	Image192              string `json:"image_192"`
	Image512              string `json:"image_512"`
	ImageOriginal         string `json:"image_original"`
	Title                 string `json:"title"`
	BotID                 string `json:"bot_id,omitempty"`
	ApiAppID              string `json:"api_app_id,omitempty"`
	StatusText            string `json:"status_text,omitempty"`
	StatusEmoji           string `json:"status_emoji,omitempty"`
	StatusExpiration      int    `json:"status_expiration"`
	Team                  string `json:"team"`
}

func (s *SlackClient) GetSlackMemberEmail(i string) (*SlackUserInfo, error) {
	var resp SlackUserInfo

	err := s.user(http.MethodGet, fmt.Sprintf("info?user=%s", i), &resp)

	return &resp, err
}

func (s *SlackClient) GetSlackID(i string) (*SlackUserInfo, error) {
	var resp SlackUserInfo

	err := s.user(http.MethodGet, fmt.Sprintf("lookupByEmail?email=%s", i), &resp)

	return &resp, err

}

func (p *UserPagination) list() (*ListResponse, error) {
	var resp ListResponse

	err := p.c.user(http.MethodGet, fmt.Sprintf("list?limit=%d&cursor=%s", p.limit, p.cursor), &resp)

	return &resp, err
}

func (s *SlackClient) ListUsers() []Member {
	p := &UserPagination{
		users:  []Member{},
		limit:  200,
		cursor: "",
		c:      s,
	}

	for {
		resp, err := p.list()
		if err != nil {
			s.log.Err(err).Msg("error listing users")
			break
		}
		if resp.ResponseMetadata.NextCursor == "" {
			break
		}
		p.cursor = resp.ResponseMetadata.NextCursor
		p.users = append(p.users, resp.Members...)
	}

	return p.users
}

func (s *SlackClient) user(m, q string, i interface{}) error {
	req, err := s.newRequest(m, fmt.Sprintf("users.%s", q), nil)
	if err != nil {
		s.log.Err(err).Msg("error building request")
		return err
	}

	err = s.do(req, i)
	if err != nil {
		s.log.Err(err).Msg("error sending request")
		return err
	}

	return nil
}
