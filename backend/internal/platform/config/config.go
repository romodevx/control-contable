// Package config loads the process configuration from environment variables
// and secrets from a SecretSource port, validating everything up front and
// reporting every problem at once.
//
// Variables: APP_ENV (required: dev, staging, production), PORT (default
// 8080, as set by Cloud Run), LOG_LEVEL (debug, info, warn, error; default
// info), RATE_LIMIT_RPS and RATE_LIMIT_BURST (per-client-IP token bucket;
// default 10 and 20), TRUSTED_PROXY_HOPS (reverse proxies in front of the
// server that append to X-Forwarded-For; 0 trusts only the TCP peer address;
// defaults to 0 in dev and is required in staging and production). Secrets:
// DATABASE_URL.
package config

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
)

// Lookup reads one environment variable; os.LookupEnv in production.
type Lookup func(key string) (string, bool)

// Environment is the deployment environment.
type Environment string

// The deployment environments.
const (
	EnvDev        Environment = "dev"
	EnvStaging    Environment = "staging"
	EnvProduction Environment = "production"
)

var environments = []Environment{EnvDev, EnvStaging, EnvProduction}

// Variable and secret names.
const (
	keyAppEnv      = "APP_ENV"
	keyPort        = "PORT"
	keyLogLevel    = "LOG_LEVEL"
	keyRateRPS     = "RATE_LIMIT_RPS"
	keyRateBurst   = "RATE_LIMIT_BURST"
	keyProxyHops   = "TRUSTED_PROXY_HOPS"
	secretDatabase = "DATABASE_URL"
)

const (
	defaultPort = "8080"
	minPort     = 1
	maxPort     = 65535

	defaultRateRPS   = 10
	defaultRateBurst = 20
	maxRate          = 10000 // bounds both RPS and burst

	defaultProxyHops = 0
	maxProxyHops     = 10
)

// RateLimit configures the per-client token bucket: PerSecond tokens are
// refilled each second, up to Burst.
type RateLimit struct {
	PerSecond int
	Burst     int
}

// Config is the validated process configuration.
type Config struct {
	Env         Environment
	HTTPAddr    string // listen address, ":<PORT>"
	LogLevel    slog.Level
	DatabaseURL Secret
	RateLimit   RateLimit
	// TrustedProxyHops is how many reverse proxies append to X-Forwarded-For
	// in front of the server; 0 trusts only the TCP peer address.
	TrustedProxyHops int
}

// Load reads and validates the configuration. The error joins every problem
// found, and never contains a secret value.
func Load(ctx context.Context, lookup Lookup, secrets SecretSource) (Config, error) {
	var (
		cfg  Config
		errs []error
		err  error
	)
	if cfg.Env, err = parseEnvironment(lookup); err != nil {
		errs = append(errs, err)
	}
	if cfg.HTTPAddr, err = parseHTTPAddr(lookup); err != nil {
		errs = append(errs, err)
	}
	if cfg.LogLevel, err = parseLogLevel(lookup); err != nil {
		errs = append(errs, err)
	}
	if cfg.RateLimit.PerSecond, err = parseIntInRange(lookup, keyRateRPS, defaultRateRPS, 1, maxRate); err != nil {
		errs = append(errs, err)
	}
	if cfg.RateLimit.Burst, err = parseIntInRange(lookup, keyRateBurst, defaultRateBurst, 1, maxRate); err != nil {
		errs = append(errs, err)
	}
	if cfg.TrustedProxyHops, err = parseProxyHops(lookup, cfg.Env); err != nil {
		errs = append(errs, err)
	}
	if cfg.DatabaseURL, err = secrets.Secret(ctx, secretDatabase); err != nil {
		errs = append(errs, fmt.Errorf("secret %s: %w", secretDatabase, err))
	}
	if len(errs) > 0 {
		return Config{}, fmt.Errorf("config: %w", errors.Join(errs...))
	}
	return cfg, nil
}

func parseEnvironment(lookup Lookup) (Environment, error) {
	raw, ok := lookup(keyAppEnv)
	if !ok || raw == "" {
		return "", fmt.Errorf("%s is required", keyAppEnv)
	}
	env := Environment(raw)
	if !slices.Contains(environments, env) {
		return "", fmt.Errorf("%s %q is not one of dev, staging, production", keyAppEnv, raw)
	}
	return env, nil
}

func parseHTTPAddr(lookup Lookup) (string, error) {
	port := valueOr(lookup, keyPort, defaultPort)
	n, err := strconv.Atoi(port)
	if err != nil || n < minPort || n > maxPort {
		return "", fmt.Errorf("%s %q is not a TCP port", keyPort, port)
	}
	return ":" + port, nil
}

func parseLogLevel(lookup Lookup) (slog.Level, error) {
	var level slog.Level
	raw := valueOr(lookup, keyLogLevel, slog.LevelInfo.String())
	if err := level.UnmarshalText([]byte(raw)); err != nil {
		return 0, fmt.Errorf("%s %q: %w", keyLogLevel, raw, err)
	}
	return level, nil
}

// parseProxyHops requires TRUSTED_PROXY_HOPS outside dev: behind a load
// balancer every request shares the proxy's peer address, so the dev default
// of 0 would put every client in one rate-limit bucket.
func parseProxyHops(lookup Lookup, env Environment) (int, error) {
	if v, ok := lookup(keyProxyHops); (!ok || v == "") && env != EnvDev && env != "" {
		return 0, fmt.Errorf("%s is required when %s is %s", keyProxyHops, keyAppEnv, env)
	}
	return parseIntInRange(lookup, keyProxyHops, defaultProxyHops, 0, maxProxyHops)
}

func parseIntInRange(lookup Lookup, key string, fallback, minValue, maxValue int) (int, error) {
	raw := valueOr(lookup, key, strconv.Itoa(fallback))
	n, err := strconv.Atoi(raw)
	if err != nil || n < minValue || n > maxValue {
		return 0, fmt.Errorf("%s %q is not an integer in [%d, %d]", key, raw, minValue, maxValue)
	}
	return n, nil
}

func valueOr(lookup Lookup, key, fallback string) string {
	if v, ok := lookup(key); ok && v != "" {
		return v
	}
	return fallback
}
