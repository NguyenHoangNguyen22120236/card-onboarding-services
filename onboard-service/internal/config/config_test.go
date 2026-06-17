package config

import (
	"testing"
	"time"
)

func TestLoadReturnsDefaults(t *testing.T) {
	cfg := Load()

	if cfg.ServerAddress != ":8080" {
		t.Fatalf("ServerAddress = %q, want %q", cfg.ServerAddress, ":8080")
	}
	if cfg.CustomerManagementBaseURL != "http://localhost:8081" {
		t.Fatalf("CustomerManagementBaseURL = %q, want %q", cfg.CustomerManagementBaseURL, "http://localhost:8081")
	}
	if cfg.AccountManagementBaseURL != "http://localhost:8082" {
		t.Fatalf("AccountManagementBaseURL = %q, want %q", cfg.AccountManagementBaseURL, "http://localhost:8082")
	}
	if cfg.DownstreamTimeout != 5*time.Second {
		t.Fatalf("DownstreamTimeout = %s, want %s", cfg.DownstreamTimeout, 5*time.Second)
	}
	if cfg.RequestStatusTableName != "" {
		t.Fatalf("RequestStatusTableName = %q, want empty", cfg.RequestStatusTableName)
	}
	if cfg.AccountDetailsTableName != "" {
		t.Fatalf("AccountDetailsTableName = %q, want empty", cfg.AccountDetailsTableName)
	}
}

func TestLoadUsesServerAddress(t *testing.T) {
	t.Setenv("SERVER_ADDRESS", ":9090")
	t.Setenv("PORT", "8083")

	cfg := Load()

	if cfg.ServerAddress != ":9090" {
		t.Fatalf("ServerAddress = %q, want %q", cfg.ServerAddress, ":9090")
	}
}

func TestLoadUsesPort(t *testing.T) {
	t.Setenv("PORT", "8083")

	cfg := Load()

	if cfg.ServerAddress != ":8083" {
		t.Fatalf("ServerAddress = %q, want %q", cfg.ServerAddress, ":8083")
	}
}

func TestLoadUsesPrefixedPort(t *testing.T) {
	t.Setenv("PORT", ":8084")

	cfg := Load()

	if cfg.ServerAddress != ":8084" {
		t.Fatalf("ServerAddress = %q, want %q", cfg.ServerAddress, ":8084")
	}
}

func TestLoadUsesDownstreamOverrides(t *testing.T) {
	t.Setenv("CUSTOMER_MANAGEMENT_BASE_URL", "http://customer-service:9001")
	t.Setenv("ACCOUNT_MANAGEMENT_BASE_URL", "http://account-service:9002")
	t.Setenv("DOWNSTREAM_TIMEOUT", "250ms")
	t.Setenv("REQUEST_STATUS_TABLE_NAME", "request-status")
	t.Setenv("ACCOUNT_DETAILS_TABLE_NAME", "account-details")

	cfg := Load()

	if cfg.CustomerManagementBaseURL != "http://customer-service:9001" {
		t.Fatalf("CustomerManagementBaseURL = %q, want %q", cfg.CustomerManagementBaseURL, "http://customer-service:9001")
	}
	if cfg.AccountManagementBaseURL != "http://account-service:9002" {
		t.Fatalf("AccountManagementBaseURL = %q, want %q", cfg.AccountManagementBaseURL, "http://account-service:9002")
	}
	if cfg.DownstreamTimeout != 250*time.Millisecond {
		t.Fatalf("DownstreamTimeout = %s, want %s", cfg.DownstreamTimeout, 250*time.Millisecond)
	}
	if cfg.RequestStatusTableName != "request-status" {
		t.Fatalf("RequestStatusTableName = %q, want %q", cfg.RequestStatusTableName, "request-status")
	}
	if cfg.AccountDetailsTableName != "account-details" {
		t.Fatalf("AccountDetailsTableName = %q, want %q", cfg.AccountDetailsTableName, "account-details")
	}
}

func TestLoadIgnoresInvalidDownstreamTimeout(t *testing.T) {
	t.Setenv("DOWNSTREAM_TIMEOUT", "soon")

	cfg := Load()

	if cfg.DownstreamTimeout != 5*time.Second {
		t.Fatalf("DownstreamTimeout = %s, want %s", cfg.DownstreamTimeout, 5*time.Second)
	}
}
