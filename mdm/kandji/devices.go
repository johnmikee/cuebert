package kandji

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/johnmikee/cuebert/mdm"
)

// DeviceResults is a list of DeviceResult
type DeviceResults []DeviceResult

func unmarshalDeviceResults(data []byte) (DeviceResults, error) {
	var r DeviceResults
	err := json.Unmarshal(data, &r)
	return r, err
}

func unmarshalDeviceDetails(data []byte) (DeviceDetails, error) {
	var r DeviceDetails
	err := json.Unmarshal(data, &r)
	return r, err
}

// GetDeviceDetails returns details on the device
func (c *Config) GetDeviceDetails(d string) (*mdm.Device, error) {
	details, err := c.deviceDetails(d)
	if err != nil {
		return nil, err
	}

	return &mdm.Device{
		DeviceID:        details.General.DeviceID,
		DeviceName:      details.General.DeviceName,
		Model:           details.General.Model,
		SerialNumber:    details.HardwareOverview.SerialNumber,
		Platform:        details.General.Platform,
		OSVersion:       details.General.OSVersion,
		LastCheckIn:     details.Mdm.LastCheckIn,
		AssetTag:        details.General.AssetTag,
		FirstEnrollment: details.General.FirstEnrollment,
		LastEnrollment:  details.General.LastEnrollment,
		User: mdm.User{
			Email: details.Users.SystemUsers[0].Username,
			Name:  *details.Users.SystemUsers[0].Name,
			ID:    details.Users.SystemUsers[0].Uid,
		},
	}, nil
}

func (c *Config) deviceDetails(d string) (*DeviceDetails, error) {
	u := fmt.Sprintf("devices/%s/details", d)

	req, err := c.newRequest(http.MethodGet, u, nil)
	if err != nil {
		c.log.Err(err).Msg("error building request")
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		c.log.Err(err).Msg("error making request")
		return nil, err
	}

	details, err := handleDeviceDetailResponse(resp)
	if err != nil {
		return nil, err
	}

	return details, err
}

func (c *Config) list(limit, offset int) (DeviceResults, error) {
	url := fmt.Sprintf("devices?limit=%d&offset=%d", limit, offset)

	req, err := c.newRequest(http.MethodGet, url, nil)
	if err != nil {
		c.log.Err(err).Msg("error building request")
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		c.log.Info().Str("url", req.URL.String()).Msg("error making request")
		c.log.Err(err).Msg("error making request")
		return nil, err
	}

	deviceResults, err := handleDeviceUserResponse(resp)
	if err != nil {
		return nil, err
	}

	return deviceResults, nil
}

func (c *Config) listDevices(limit, offset int) (mdm.DeviceResults, error) {
	dev, err := c.list(limit, offset)
	if err != nil {
		return nil, err
	}

	results := transform(dev)

	return results, nil
}

// ListAllDevices will paginate through devices until there is no response and return the results
func (c *Config) ListAllDevices() (mdm.DeviceResults, error) {
	opts := &offsetRange{
		Limit:  300,
		Offset: 0,
	}

	count := 0
	res := mdm.DeviceResults{}
	for {
		results, err := c.listDevices(opts.Limit, opts.Offset)
		if err != nil {
			c.log.Err(err).Msg("error listing devices")
			return results, err
		}
		count += len(results)
		opts.Offset += opts.Limit
		if len(results) == 0 {
			break
		}
		res = append(res, results...)

	}

	return res, nil
}

// QueryDevices will return  devices based off query passed.
func (c *Config) QueryDevices(opts *mdm.QueryOpts) (mdm.DeviceResults, error) {
	query := fmt.Sprintf(
		"devices?asset_tag=%s&device_id=%s&device_name=%s&model=%s&os_version=%s&serial_number=%s&platform=%s&user_email=%s&user_id=%s&user_name=%s&",
		opts.AssetTag,
		opts.DeviceID,
		opts.DeviceName,
		opts.Model,
		opts.OSVersion,
		opts.SerialNumber,
		opts.Platform,
		opts.UserEmail,
		opts.UserID,
		opts.UserName,
	)

	m := regexp.MustCompile(`[a-z\_]+\=&`)
	query = m.ReplaceAllString(query, "")
	query = strings.TrimSuffix(query, "&")

	req, err := c.newRequest(http.MethodGet, query, nil)
	if err != nil {
		c.log.Err(err).Msg("error building request")
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		c.log.Err(err).Msg("error making request")
		return nil, err
	}

	deviceResults, err := handleDeviceUserResponse(resp)
	if err != nil {
		return nil, err
	}

	return transform(deviceResults), nil
}
