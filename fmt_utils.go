package tron

import (
	"strings"
)

func SplitPort(addr string) string {
	hosts := strings.SplitN(addr, ":", 2)
	if len(hosts) < 2 {
		return ""
	}
	return hosts[1]
}
