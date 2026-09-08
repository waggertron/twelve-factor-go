package config

import (
	"strings"
	"testing"
)

func validConfig() Config {
	return Config{AppHost: "0.0.0.0", DatabaseURL: "postgresql://orders:local_only@localhost/orders", RedisURL: "redis://localhost:6379", Port: 3103, WorkerConcurrency: 4, ShutdownGraceMS: 10000, ReleaseID: "test", TelemetryMode: "console"}
}

func TestSharedConfigBounds(t *testing.T) {
	for _, mode := range []string{"console", "memory", "disabled"} {
		cfg := validConfig()
		cfg.TelemetryMode = mode
		if err := Validate(cfg, "web"); err != nil {
			t.Fatalf("mode %q: %v", mode, err)
		}
	}
	for name, mutate := range map[string]func(*Config){
		"empty app host":    func(cfg *Config) { cfg.AppHost = "" },
		"privileged port":   func(cfg *Config) { cfg.Port = 1023 },
		"missing release":   func(cfg *Config) { cfg.ReleaseID = "" },
		"invalid release":   func(cfg *Config) { cfg.ReleaseID = "bad release" },
		"invalid telemetry": func(cfg *Config) { cfg.TelemetryMode = "remote" },
	} {
		cfg := validConfig()
		mutate(&cfg)
		if err := Validate(cfg, "web"); err == nil {
			t.Fatalf("%s should fail", name)
		}
	}
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
