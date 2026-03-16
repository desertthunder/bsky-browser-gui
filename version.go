package main

import "strings"

var version = "dev"

func appVersion() string {
	v := strings.TrimSpace(version)
	if v == "" {
		return "dev"
	}

	return v
}
