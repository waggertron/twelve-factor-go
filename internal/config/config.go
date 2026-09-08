package config

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	DatabaseURL       string `env:"DATABASE_URL,required"`
	RedisURL          string `env:"REDIS_URL"`
	Port              int    `env:"PORT" envDefault:"3103"`
	WorkerConcurrency int    `env:"WORKER_CONCURRENCY" envDefault:"4"`
	ShutdownGraceMS   int    `env:"SHUTDOWN_GRACE_MS" envDefault:"10000"`
	ReleaseID         string `env:"RELEASE_ID" envDefault:"dev"`
	TelemetryMode     string `env:"TELEMETRY_MODE" envDefault:"console"`
}

type Error struct{ Category string }

func (e Error) Error() string { return "invalid configuration: " + e.Category }

func Load(processType string) (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, Error{Category: "missing_or_invalid_value"}
	}
	if err := Validate(cfg, processType); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Validate(cfg Config, processType string) error {
	if err := requireScheme(cfg.DatabaseURL, "postgres", "postgresql"); err != nil {
		return Error{Category: "database_url"}
	}
	if processType == "web" || processType == "worker" {
		if err := requireScheme(cfg.RedisURL, "redis", "rediss"); err != nil {
			return Error{Category: "redis_url"}
		}
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return Error{Category: "port"}
	}
	if cfg.WorkerConcurrency < 1 || cfg.WorkerConcurrency > 32 {
		return Error{Category: "worker_concurrency"}
	}
	if cfg.ShutdownGraceMS < 1000 || cfg.ShutdownGraceMS > 60000 {
		return Error{Category: "shutdown_grace_ms"}
	}
	if cfg.TelemetryMode != "console" {
		return Error{Category: "telemetry_mode"}
	}
	return nil
}

func requireScheme(raw string, allowed ...string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return errors.New("invalid URL")
	}
	for _, scheme := range allowed {
		if parsed.Scheme == scheme {
			return nil
		}
	}
	return fmt.Errorf("unsupported scheme")
}
