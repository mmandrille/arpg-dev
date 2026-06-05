// Package config loads server configuration from the environment with
// development-friendly defaults. All settings are read once at startup.
package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all runtime settings for the server.
type Config struct {
	// Addr is the listen address, e.g. ":8080".
	Addr string
	// DatabaseURL is the Postgres connection string.
	DatabaseURL string
	// DevToken is the shared secret accepted by POST /v0/auth/dev-login.
	DevToken string
	// DebugToken is required (as X-Debug-Token) for debug-gated routes.
	DebugToken string
	// Env is "local" or "remote". Remote deployments enforce stricter rules.
	Env string
	// MetricsEnabled gates the /metrics route (deployment-gated per spec).
	MetricsEnabled bool
}

const (
	defaultAddr        = ":8080"
	defaultDatabaseURL = "postgres://arpg:arpg@localhost:5432/arpg?sslmode=disable"
	defaultDevToken    = "local-dev-token"
	defaultDebugToken  = "local-debug-token"
	defaultEnv         = "local"
)

// Load reads configuration from ARPG_* environment variables.
func Load() Config {
	return Config{
		Addr:           getenv("ARPG_ADDR", defaultAddr),
		DatabaseURL:    getenv("ARPG_DATABASE_URL", defaultDatabaseURL),
		DevToken:       getenv("ARPG_DEV_TOKEN", defaultDevToken),
		DebugToken:     getenv("ARPG_DEBUG_TOKEN", defaultDebugToken),
		Env:            getenv("ARPG_ENV", defaultEnv),
		MetricsEnabled: getenvBool("ARPG_METRICS_ENABLED", true),
	}
}

// IsLocal reports whether the server is running in local development mode.
func (c Config) IsLocal() bool { return c.Env == "local" }

// Validate enforces invariants that must hold before the server starts.
// Remote deployments must not run with the default debug token, which would
// expose debug/inspection routes with a publicly known secret.
func (c Config) Validate() error {
	if c.Addr == "" {
		return fmt.Errorf("config: ARPG_ADDR must not be empty")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("config: ARPG_DATABASE_URL must not be empty")
	}
	if !c.IsLocal() {
		if c.DebugToken == "" || c.DebugToken == defaultDebugToken {
			return fmt.Errorf("config: ARPG_DEBUG_TOKEN must be set to a non-default value when ARPG_ENV=%q", c.Env)
		}
		if c.DevToken == "" || c.DevToken == defaultDevToken {
			return fmt.Errorf("config: ARPG_DEV_TOKEN must be set to a non-default value when ARPG_ENV=%q", c.Env)
		}
	}
	return nil
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getenvBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
