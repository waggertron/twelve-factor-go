package config

import (
	"strings"
	"testing"
)

func validConfig() Config {
	return Config{DatabaseURL: "postgresql://orders:local_only@localhost/orders", RedisURL: "redis://localhost:6379", Port: 3103, WorkerConcurrency: 4, ShutdownGraceMS: 10000, ReleaseID: "test", TelemetryMode: "console"}
}

func TestValidateConfig(t *testing.T) {
	if err := Validate(validConfig(), "web"); err != nil {
		t.Fatal(err)
	}
	cfg := validConfig()
	cfg.WorkerConcurrency = 33
	if err := Validate(cfg, "worker"); err == nil {
		t.Fatal("unbounded concurrency should fail")
	}
}

func TestConfigErrorDoesNotLeakValue(t *testing.T) {
	cfg := validConfig()
	cfg.DatabaseURL = "not-a-database-url"
	err := Validate(cfg, "admin")
	if err == nil || strings.Contains(err.Error(), cfg.DatabaseURL) {
		t.Fatalf("expected safe categorized error, got %v", err)
	}
}
