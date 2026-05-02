package config

import "testing"

func TestLoadRejectsInvalidIPRateLimit(t *testing.T) {
	t.Setenv("RATE_LIMIT_IP_RATE", "0")
	t.Setenv("RATE_LIMIT_IP_BURST", "20")
	t.Setenv("RATE_LIMIT_USER_RATE", "200")
	t.Setenv("RATE_LIMIT_USER_BURST", "40")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid IP rate config")
	}
}

func TestLoadRejectsInvalidIPBurst(t *testing.T) {
	t.Setenv("RATE_LIMIT_IP_RATE", "100")
	t.Setenv("RATE_LIMIT_IP_BURST", "-1")
	t.Setenv("RATE_LIMIT_USER_RATE", "200")
	t.Setenv("RATE_LIMIT_USER_BURST", "40")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid IP burst config")
	}
}

func TestLoadRejectsInvalidUserRateLimit(t *testing.T) {
	t.Setenv("RATE_LIMIT_IP_RATE", "100")
	t.Setenv("RATE_LIMIT_IP_BURST", "20")
	t.Setenv("RATE_LIMIT_USER_RATE", "0")
	t.Setenv("RATE_LIMIT_USER_BURST", "40")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid user rate config")
	}
}

func TestLoadRejectsInvalidUserBurst(t *testing.T) {
	t.Setenv("RATE_LIMIT_IP_RATE", "100")
	t.Setenv("RATE_LIMIT_IP_BURST", "20")
	t.Setenv("RATE_LIMIT_USER_RATE", "200")
	t.Setenv("RATE_LIMIT_USER_BURST", "0")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid user burst config")
	}
}

func TestLoadAllowsValidRateLimits(t *testing.T) {
	t.Setenv("RATE_LIMIT_IP_RATE", "100")
	t.Setenv("RATE_LIMIT_IP_BURST", "20")
	t.Setenv("RATE_LIMIT_USER_RATE", "200")
	t.Setenv("RATE_LIMIT_USER_BURST", "40")

	_, err := Load()
	if err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}
