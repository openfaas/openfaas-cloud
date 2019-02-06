package main

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// RouterConfig configuration for the router
type RouterConfig struct {
	Port        string
	UpstreamURL string
	AuthURL     string
	Timeout     time.Duration
}

// NewRouterConfig create a new RouterConfig by loading
// config from environmental variables.
func NewRouterConfig() RouterConfig {
	cfg := RouterConfig{
		Port: "8080",
	}

	if portVal, exists := os.LookupEnv("port"); exists && len(portVal) > 0 {
		cfg.Port = portVal
	}

	if val, exists := os.LookupEnv("auth_url"); exists && len(val) > 0 {
		if strings.HasSuffix(val, "/") == false {
			val = val + "/"
		}

		cfg.AuthURL = val
	}

	if up, exists := os.LookupEnv("upstream_url"); exists && len(up) > 0 {
		if strings.HasSuffix(up, "/") == false {
			up = up + "/"
		}

		cfg.UpstreamURL = up
	}

	cfg.Timeout = parseIntOrDurationValue(os.Getenv("timeout"), time.Second*60)

	return cfg
}

func parseIntOrDurationValue(val string, fallback time.Duration) time.Duration {
	if len(val) > 0 {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr == nil && parsedVal >= 0 {
			return time.Duration(parsedVal) * time.Second
		}
	}

	duration, durationErr := time.ParseDuration(val)
	if durationErr != nil {
		return fallback
	}
	return duration
}
