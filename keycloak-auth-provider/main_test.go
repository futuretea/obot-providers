package main

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/oauth2-proxy/oauth2-proxy/v7/pkg/apis/options"
	"github.com/obot-platform/providers/keycloak-auth-provider/pkg/config"
)

func TestManifestDeclaresGroupIDPrefix(t *testing.T) {
	body, err := os.ReadFile("../auth-providers/keycloak-auth-provider.yaml")
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	text := string(body)
	quoted := `groupIDPrefix: "` + config.GroupIDPrefix + `"`
	bare := "groupIDPrefix: " + config.GroupIDPrefix
	if !strings.Contains(text, quoted) && !strings.Contains(text, bare) {
		t.Fatalf("manifest missing %q or %q", quoted, bare)
	}
}

func TestConfigureCookieSetsCSRFExpireToThirtyMinutes(t *testing.T) {
	opts := options.NewOptions()
	cfg := &config.Config{
		CookieSecret:         []byte("0123456789abcdef0123456789abcdef"),
		ObotServerURL:        "http://localhost:8080",
		TokenRefreshDuration: time.Hour,
	}

	configureCookie(opts, cfg)

	if opts.Cookie.CSRFExpire != 30*time.Minute {
		t.Fatalf("expected CSRFExpire to be 30 minutes, got %s", opts.Cookie.CSRFExpire)
	}
}
