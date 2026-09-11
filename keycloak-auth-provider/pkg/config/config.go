package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/obot-platform/providers/auth-providers-common/pkg/env"
)

// Environment variable names
const (
	envPort               = "PORT"
	envListenHost         = "OBOT_PROVIDER_LISTEN_HOST"
	envIssuerURL          = "OAUTH2_PROXY_OIDC_ISSUER_URL"
	envDebug              = "OBOT_AUTH_DEBUG"
	envGroupSearchEnabled = "OBOT_KEYCLOAK_GROUP_SEARCH_ENABLED"
)

// Default values
const (
	defaultPort            = "9999"
	defaultListenHost      = "127.0.0.1"
	userInfoEndpointSuffix = "/protocol/openid-connect/userinfo"
	// GroupIDPrefix is the host-required group ID namespace. Must match the
	// keycloak-auth-provider.yaml groupIDPrefix and auth.ValidateGroupIDPrefix.
	GroupIDPrefix = "keycloak/"
)

// GroupID maps a Keycloak group path onto the declared host namespace.
func GroupID(path string) string {
	return GroupIDPrefix + strings.TrimPrefix(path, "/")
}

// ErrMissingIssuerURL indicates the OIDC issuer URL is not configured
var ErrMissingIssuerURL = errors.New(envIssuerURL + " is required but not set")

// Options contains raw configuration from environment variables
type Options struct {
	ClientID                          string `env:"OBOT_KEYCLOAK_AUTH_PROVIDER_CLIENT_ID"`
	ClientSecret                      string `env:"OBOT_KEYCLOAK_AUTH_PROVIDER_CLIENT_SECRET"`
	ObotServerURL                     string `env:"OBOT_SERVER_PUBLIC_URL,OBOT_SERVER_URL"`
	PostgresConnectionDSN             string `env:"OBOT_AUTH_PROVIDER_POSTGRES_CONNECTION_DSN" optional:"true"`
	PostgresMaxConnections            int    `env:"OBOT_AUTH_PROVIDER_POSTGRES_MAX_CONNECTIONS" optional:"true"`
	PostgresMaxIdleConnections        int    `env:"OBOT_AUTH_PROVIDER_POSTGRES_MAX_IDLE_CONNECTIONS" optional:"true"`
	PostgresConnectionLifetimeSeconds int    `env:"OBOT_AUTH_PROVIDER_POSTGRES_CONNECTION_LIFETIME_SECONDS" optional:"true"`
	AuthCookieSecret                  string `env:"OBOT_AUTH_PROVIDER_COOKIE_SECRET"`
	AuthEmailDomains                  string `env:"OBOT_AUTH_PROVIDER_EMAIL_DOMAINS" default:"*"`
	AuthTokenRefreshDuration          string `env:"OBOT_AUTH_PROVIDER_TOKEN_REFRESH_DURATION" default:"1h" optional:"true"`
	LoggingEnabled                    string `env:"OBOT_AUTH_PROVIDER_ENABLE_LOGGING" optional:"true"`
}

// Config is the parsed and validated configuration
type Config struct {
	ClientID                          string
	ClientSecret                      string
	ObotServerURL                     string
	PostgresConnectionDSN             string
	PostgresMaxConnections            int
	PostgresMaxIdleConnections        int
	PostgresConnectionLifetimeSeconds int
	CookieSecret                      []byte
	EmailDomains                      []string
	TokenRefreshDuration              time.Duration
	GroupSearchEnabled                bool
	Debug                             bool
	LoggingEnabled                    bool
	Port                              string
	ListenHost                        string
	IssuerURL                         string
}

// LoadFromEnv loads and validates configuration from environment variables
func LoadFromEnv() (*Config, error) {
	var opts Options
	if err := env.LoadEnvForStruct(&opts); err != nil {
		return nil, fmt.Errorf("load options: %w", err)
	}

	refreshDuration, err := time.ParseDuration(opts.AuthTokenRefreshDuration)
	if err != nil {
		return nil, fmt.Errorf("parse token refresh duration %q: %w", opts.AuthTokenRefreshDuration, err)
	}
	if refreshDuration < 0 {
		return nil, errors.New("token refresh duration must be positive")
	}

	cookieSecret, err := base64.StdEncoding.DecodeString(opts.AuthCookieSecret)
	if err != nil {
		return nil, fmt.Errorf("decode cookie secret: %w", err)
	}

	issuerURL := os.Getenv(envIssuerURL)
	if issuerURL == "" {
		return nil, ErrMissingIssuerURL
	}

	return &Config{
		ClientID:                          opts.ClientID,
		ClientSecret:                      opts.ClientSecret,
		ObotServerURL:                     opts.ObotServerURL,
		PostgresConnectionDSN:             opts.PostgresConnectionDSN,
		PostgresMaxConnections:            opts.PostgresMaxConnections,
		PostgresMaxIdleConnections:        opts.PostgresMaxIdleConnections,
		PostgresConnectionLifetimeSeconds: opts.PostgresConnectionLifetimeSeconds,
		CookieSecret:                      cookieSecret,
		EmailDomains:                      parseEmailDomains(opts.AuthEmailDomains),
		TokenRefreshDuration:              refreshDuration,
		GroupSearchEnabled:                os.Getenv(envGroupSearchEnabled) != "false",
		Debug:                             os.Getenv(envDebug) == "true",
		LoggingEnabled:                    strings.EqualFold(opts.LoggingEnabled, "true"),
		Port:                              getEnvOrDefault(envPort, defaultPort),
		ListenHost:                        getEnvOrDefault(envListenHost, defaultListenHost),
		IssuerURL:                         issuerURL,
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func parseEmailDomains(domains string) []string {
	if domains == "" {
		return nil
	}
	parts := strings.Split(domains, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// UserInfoURL returns the OIDC UserInfo endpoint URL
func (c *Config) UserInfoURL() string {
	return strings.TrimRight(c.IssuerURL, "/") + userInfoEndpointSuffix
}
