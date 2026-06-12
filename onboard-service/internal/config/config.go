package config

import (
	"os"
	"strings"
	"time"
)

const (
	defaultServerAddress             = ":8080"
	defaultCustomerManagementBaseURL = "http://localhost:8081"
	defaultAccountManagementBaseURL  = "http://localhost:8082"
	defaultDownstreamTimeout         = 5 * time.Second
)

type Config struct {
	ServerAddress             string
	CustomerManagementBaseURL string
	AccountManagementBaseURL  string
	DownstreamTimeout         time.Duration
}

func Load() Config {
	return Config{
		ServerAddress:             serverAddress(),
		CustomerManagementBaseURL: envOrDefault("CUSTOMER_MANAGEMENT_BASE_URL", defaultCustomerManagementBaseURL),
		AccountManagementBaseURL:  envOrDefault("ACCOUNT_MANAGEMENT_BASE_URL", defaultAccountManagementBaseURL),
		DownstreamTimeout:         durationEnvOrDefault("DOWNSTREAM_TIMEOUT", defaultDownstreamTimeout),
	}
}

func serverAddress() string {
	if value := strings.TrimSpace(os.Getenv("SERVER_ADDRESS")); value != "" {
		return value
	}

	if value := strings.TrimSpace(os.Getenv("PORT")); value != "" {
		if strings.HasPrefix(value, ":") {
			return value
		}
		return ":" + value
	}

	return defaultServerAddress
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func durationEnvOrDefault(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return duration
}
