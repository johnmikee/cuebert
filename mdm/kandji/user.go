/*
The code below is admittedly convoluted.

There are some oddities in the serializers that return the user data in the response from Kandji.
If no user is reported for the device the response is returned as "" instead of an empty map/dictionary.

This presents a challenge when unmarshalling the user response into the DeviceResult struct. If the response
for a device with no user was an empty map we would not be facing a problem - hopefully that will change in the future.

It is entirely possible to streamline what is happening below. This was such a crucial bit of information needed for cue
that when I found a solution via https://github.com/quicktype/quicktype I cleaned it up a bit and moved on.

Below is an example of what I used to generate the structs for the DeviceResult and DeviceDetail as it pertained to the
user response. To generate this you must first curl the endpoint to get the response for quicktype to transform for you.

DeviceResults:

 1. curl --location --request GET 'https://$subdomain.clients.us-1.kandji.io/api/v1/devices' \
    --header 'Authorization: Bearer $token' >> deviceresult.json

 2. quicktype \
    --src deviceresult.json \
    --src-lang json \
    --lang go \
    --top-level DeviceResults \
    --out device_result.go

DeviceDetails:

 1. Grab the device ID of any device

 2. curl --location --request GET 'https://$subdomain.clients.us-1.kandji.io/api/v1/devices/$deviceID/details' \
    --header 'Authorization: Bearer $token' >> devicedetails.json

 3. quicktype \
    --src devicedetails.json \
    --src-lang json \
    --lang go \
    --top-level DeviceDetails \
    --out device_detail.go

To access the user information you must check if the response was nil if you do not want the program to panic.
Example:

	k := kandji.NewClient(c.MDMKey, c.MDMURL, nil, &c.log)
	dev, err := k.ListAllDevices()
	if err != nil {
		return nil, err
	}
	machines := []devices.DeviceInfo{}

	for _, d := range dev {
		di := devices.DeviceInfo{
			DeviceID:     d.DeviceID,
			DeviceName:   d.DeviceName,
			Model:        d.Model,
			SerialNumber: d.SerialNumber,
			Platform:     string(d.Platform),
			OSVersion:    d.OSVersion,
			LastCheckIn:  d.LastCheckIn,
			User:         "",
		}

		if d.User.UserClass != nil {
			di.User = d.User.UserClass.Email
		}

		machines = append(machines, di)
	}
*/
package kandji

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/johnmikee/cuebert/mdm"
)

// UserUnion joins the potential types that can be returned for a user.
type UserUnion struct {
	String    *string
	UserClass *UserClass
}

// UserClass is the struct that contains the user information.
// This may be nil if no user is reported for the device.
type UserClass struct {
	Email      string `json:"email"`
	Name       string `json:"name"`
	ID         int64  `json:"id"`
	IsArchived bool   `json:"is_archived"`
}

func setMDMUser(u *UserUnion) mdm.User {
	user := mdm.User{}
	if u.UserClass != nil {
		user.Email = u.UserClass.Email
		user.ID = strconv.Itoa(int(u.UserClass.ID))
		user.Name = u.UserClass.Name
	}
	return user
}

func unmarshalUnion(data []byte,
	pi **int64,
	pf **float64,
	pb **bool,
	ps **string,
	haveArray bool,
	pa interface{},
	haveObject bool,
	pc interface{},
	haveMap bool,
	pm interface{},
	haveEnum bool,
	pe interface{},
	nullable bool) (bool, error) {

	if pi != nil {
		*pi = nil
	}
	if pf != nil {
		*pf = nil
	}
	if pb != nil {
		*pb = nil
	}
	if ps != nil {
		*ps = nil
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	tok, err := dec.Token()
	if err != nil {
		return false, err
	}

	switch v := tok.(type) {
	case json.Number:
		if pi != nil {
			i, err := v.Int64()
			if err == nil {
				*pi = &i
				return false, nil
			}
		}
		if pf != nil {
			f, err := v.Float64()
			if err == nil {
				*pf = &f
				return false, nil
			}
			return false, errors.New("unparsable number")
		}
		return false, errors.New("union does not contain number")
	case float64:
		return false, errors.New("decoder should not return float64")
	case bool:
		if pb != nil {
			*pb = &v
			return false, nil
		}
		return false, errors.New("union does not contain bool")
	case string:
		if haveEnum {
			return false, json.Unmarshal(data, pe)
		}
		if ps != nil {
			*ps = &v
			return false, nil
		}
		return false, errors.New("union does not contain string")
	case nil:
		if nullable {
			return false, nil
		}
		return false, errors.New("union does not contain null")
	case json.Delim:
		if v == '{' {
			if haveObject {
				return true, json.Unmarshal(data, pc)
			}
			if haveMap {
				return false, json.Unmarshal(data, pm)
			}
			return false, errors.New("union does not contain object")
		}
		if v == '[' {
			if haveArray {
				return false, json.Unmarshal(data, pa)
			}
			return false, errors.New("union does not contain array")
		}
		return false, errors.New("cannot handle delimiter")
	}
	return false, errors.New("cannot unmarshal union")

}

func marshalUnion(
	pi *int64,
	pf *float64,
	pb *bool,
	ps *string,
	haveArray bool,
	pa interface{},
	haveObject bool,
	pc interface{},
	haveMap bool,
	pm interface{},
	haveEnum bool,
	pe interface{},
	nullable bool) ([]byte, error) {
	if pi != nil {
		return json.Marshal(*pi)
	}
	if pf != nil {
		return json.Marshal(*pf)
	}
	if pb != nil {
		return json.Marshal(*pb)
	}
	if ps != nil {
		return json.Marshal(*ps)
	}
	if haveArray {
		return json.Marshal(pa)
	}
	if haveObject {
		return json.Marshal(pc)
	}
	if haveMap {
		return json.Marshal(pm)
	}
	if haveEnum {
		return json.Marshal(pe)
	}
	if nullable {
		return json.Marshal(nil)
	}
	return nil, errors.New("union must not be null")
}

func (x *UserUnion) UnmarshalJSON(data []byte) error {
	x.UserClass = nil
	var c UserClass

	object, err := unmarshalUnion(
		data,
		nil,
		nil,
		nil,
		&x.String,
		false,
		nil,
		true,
		&c,
		false,
		nil,
		false,
		nil,
		false)

	if err != nil {
		return err
	}
	if object {
		x.UserClass = &c
	}
	return nil
}

func (x *UserUnion) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, x.String, false, nil, x.UserClass != nil, x.UserClass, false, nil, false, nil, false)
}

func handleDeviceUserResponse(resp *http.Response) (DeviceResults, error) {
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	deviceResults, err := unmarshalDeviceResults(b)
	if err != nil {
		return nil, err
	}

	return deviceResults, nil
}

func handleDeviceDetailResponse(resp *http.Response) (*DeviceDetails, error) {
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	deviceDetails, err := unmarshalDeviceDetails(b)
	if err != nil {
		return nil, err
	}

	return &deviceDetails, nil
}

// GetUsers implements mdm.Provider.
func (c *Config) GetUsers(opts *mdm.QueryOpts) ([]mdm.User, error) {
	var res mdm.DeviceResults
	var err error

	if opts != nil {
		res, err = c.withQuery(opts)
	} else {
		res, err = c.all()
	}

	if err != nil {
		return nil, err
	}

	mu := []mdm.User{}
	for i := range res {
		mu = append(mu, res[i].User)
	}

	return mu, nil
}

func (c *Config) all() (mdm.DeviceResults, error) {
	return c.ListAllDevices()
}

func (c *Config) withQuery(opts *mdm.QueryOpts) (mdm.DeviceResults, error) {
	return c.QueryDevices(opts)
}
