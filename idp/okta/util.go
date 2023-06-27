package okta

import (
	"strings"
)

func linkSorter(l []string) string {
	var link string
	for _, i := range l {
		res := strings.Split(i, ";")
		check := strings.Split(res[1], "=")
		if strings.Contains(check[1], "next") {
			link = res[0]
		}
	}
	link = strings.ReplaceAll(link, "<", "")
	link = strings.ReplaceAll(link, ">", "")

	return link
}

func userNameChecker(un, domain string) string {
	// if the org has usernames set to the email
	// validate they end in the provided domain
	if !strings.HasSuffix(un, domain) {
		// make sure an @ wasnt on there, or the wrong domain
		if strings.Contains(un, "@") {
			parts := strings.Split(un, "@")
			un = parts[0]
		}
		return un + domain
	}

	return un
}
