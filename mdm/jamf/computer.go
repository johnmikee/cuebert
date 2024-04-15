package jamf

import (
	"net/http"

	"github.com/pkg/errors"
)

// BasicComputerInfo holds basic information associated with Jamf device
type BasicComputerInfo struct {
	GeneralInformation
}

// Computer holds information associated with a Jamf device
type Computer struct {
	Info ComputerDetails `json:"computer" xml:"computer,omitempty"`
}

// ComputerDetails holds information associated with a Jamf device
type ComputerDetails struct {
	ID int `json:"id,omitempty" xml:"id,omitempty"`
}

// GeneralInformation holds basic information associated with Jamf device
type GeneralInformation struct {
	ID                  int    `json:"id,omitempty" xml:"id,omitempty"`
	Name                string `json:"name" xml:"name,omitempty"`
	MACAddress          string `json:"mac_address" xml:"mac_address,omitempty"`
	SerialNumber        string `json:"serial_number" xml:"serial_number,omitempty"`
	UDID                string `json:"udid" xml:"udid,omitempty"`
	JamfVersion         string `json:"jamf_version" xml:"jamf_version,omitempty"`
	Platform            string `json:"platform" xml:"platform,omitempty"`
	MDMCapable          bool   `json:"mdm_capable" xml:"mdm_capable,omitempty"`
	ReportDate          string `json:"report_date" xml:"report_date,omitempty"`
	IPAddress           string `json:"ip_address" xml:"ip_address,omitempty"`
	LastReportedIP      string `json:"last_reported_ip" xml:"last_reported_ip,omitempty"`
	LastEnrolledDateUTC string `json:"last_enrolled_date_utc" xml:"last_enrolled_date_utc,omitempty"`
}

func (c *Client) Computers() ([]BasicComputerInfo, error) {
	req, err := c.newRequest(http.MethodGet, "computers", false, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error building request")
	}

	var computers []BasicComputerInfo
	_, err = c.do(req, &computers)
	if err != nil {
		return nil, errors.Wrap(err, "error making request")
	}

	return computers, nil
}
