package helpers

import (
	"fmt"
	"strings"
)

// URLShaper will make sure the baseurl ends with a "/" as the suffix
// to allow the endpoint urls to be joined without issue.
//
// Additionally, if the endpoint has a specific suffix such as /api or /api/v1
// those will be appended to the base.
func URLShaper(u, suffix string) string {
	// add trailing forward slash (if missing)
	if !strings.HasSuffix(u, "/") {
		u += "/"
	}

	if suffix != "" {
		// remove leading forward slash (if present)
		suffix = strings.TrimPrefix(suffix, "/")

		// add suffix to base url
		if !strings.HasSuffix(u, suffix) {
			u += suffix
		}
	}

	return u
}

// TokenValidator will make sure the token/key used is in the format the api needs.
//
// In the context of this program all keys used follow the pattern of $prefix $token so
// this simple check will work. Should new packages be introduced this functionality may
// need to be revisited.
//
// ex: token=ABC-123 prefix=SSWS
func TokenValidator(token, prefix string) string {
	// trim leading and trailing spaces from the token
	token = strings.TrimSpace(token)

	// to rule out formatting errors remove the prefix from the
	// token string if it is present.
	if strings.HasPrefix(token, prefix+" ") {
		token = strings.TrimPrefix(token, prefix+" ")
	}

	return fmt.Sprintf("%s %s", prefix, token)
}
