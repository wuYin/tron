package tron

import (
	"strings"
	"time"
)

func port(addr string) string {
	hosts := strings.SplitN(addr, ":", 2)
	if len(hosts) < 2 {
		return ""
	}
	return hosts[1]
}

func wait(n int64) {
	time.Sleep(time.Duration(n) * time.Second)
}