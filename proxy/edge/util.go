package edge

import "strings"

func isSubDomain(domain, host string) bool {
	_, after, found := strings.Cut(host, ".")
	if !found {
		return false
	}

	return after == domain
}
