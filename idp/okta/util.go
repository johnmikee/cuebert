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
