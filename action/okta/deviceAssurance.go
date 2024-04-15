package okta

import (
	"time"

	"github.com/johnmikee/cuebert/idp/okta"
)

/*
This action sets the minimum OS version for a device assurance policy.
This will effectively block any devices that do not meet the minimum OS version from accessing apps that are protected by the policy.
*/

// SetDeviceAssuranceVersion sets the minimum OS version for a device assurance policy.
func SetDeviceAssuranceVersion(o *okta.Client, policy string, version string) (time.Time, error) {
	resp, err := o.DeviceAssurancePolicies().Update(
		okta.DAComputerPolicy{
			ID: policy,
			OsVersion: okta.OsVersion{
				Minimum: version,
			},
		},
	)
	if err != nil {
		return time.Time{}, err
	}
	return resp.LastUpdate, nil
}
