package kandji

import (
	"fmt"
	"time"

	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/pkg/requester"
)

// DeviceResults is a list of DeviceResult.
type DeviceResults []DeviceResult

// GetDeviceDetails returns details on the device.
func (c *Client) GetDeviceDetails(d string) (*mdm.Device, error) {
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
		LastCheckIn:     kandjiTimeFixer(details.Mdm.LastCheckIn),
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

func kandjiTimeFixer(t string) *time.Time {
	layout := "2006-01-02 15:04:05.000000-07:00"

	ts, err := time.Parse(layout, t)
	if err != nil {
		return nil
	}

	return &ts
}

func (c *Client) deviceDetails(d string) (*DeviceDetails, error) {
	u := fmt.Sprintf("devices/%s/details", d)

	req, err := c.newRequest(u)
	if err != nil {
		c.log.Debug().AnErr("error", err).Msg("building request")
		return nil, err
	}

	var deviceDetails DeviceDetails
	_, err = c.do(req, &deviceDetails)
	if err != nil {
		c.log.Debug().AnErr("error", err).Msg("making request")
		return nil, err
	}

	return &deviceDetails, nil
}

func (c *Client) getDevice(d string) (*DeviceGet, error) {
	u := fmt.Sprintf("devices/%s", d)

	req, err := c.newRequest(u)
	if err != nil {
		c.log.Debug().AnErr("error", err).Msg("building request")
		return nil, err
	}

	var deviceDetails DeviceGet
	_, err = c.do(req, &deviceDetails)
	if err != nil {
		c.log.Debug().AnErr("error", err).Msg("making request")
		return nil, err
	}

	return &deviceDetails, nil
}

func (c *Client) list(limit, offset int) (DeviceResults, error) {
	url := fmt.Sprintf("devices?limit=%d&offset=%d", limit, offset)

	req, err := c.newRequest(url)
	if err != nil {
		c.log.Debug().AnErr("error", err).Msg("building request")
		return nil, err
	}

	var res DeviceResults
	_, err = c.do(req, &res)
	if err != nil {
		c.log.Debug().
			AnErr("err", err).
			Str("url", req.URL.String()).
			Msg("error making request")
		return nil, err
	}

	return res, nil
}

func (c *Client) listDevices(limit, offset int) (mdm.DeviceResults, error) {
	dev, err := c.list(limit, offset)
	if err != nil {
		return nil, err
	}

	results := transform(dev)

	return results, nil
}

// ListAllDevices will paginate through devices until there is no response and return the results.
func (c *Client) ListAllDevices() (mdm.DeviceResults, error) {
	opts := &offsetRange{
		Limit:  300,
		Offset: 0,
	}

	count := 0
	res := mdm.DeviceResults{}
	for {
		results, err := c.listDevices(opts.Limit, opts.Offset)
		if err != nil {
			c.log.Info().AnErr("error", err).Msg("listing devices")
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
func (c *Client) QueryDevices(opts *mdm.QueryOpts) (mdm.DeviceResults, error) {
	query, err := requester.BuildQuery(
		&requester.Query{
			Endpoint: "devices",
			Params:   requester.StructToQuery(opts),
		},
	)
	if err != nil {
		c.log.Debug().AnErr("error", err).Msg("building query")
		return nil, err
	}
	req, err := c.newRequest(query)
	if err != nil {
		c.log.Debug().AnErr("error", err).Msg("building request")
		return nil, err
	}

	var deviceResults DeviceResults
	_, err = c.do(req, &deviceResults)
	if err != nil {
		c.log.Debug().AnErr("error", err).Msg("making request")
		return nil, err
	}

	return transform(deviceResults), nil
}
