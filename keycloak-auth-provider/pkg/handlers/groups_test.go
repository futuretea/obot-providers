package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/obot-platform/providers/keycloak-auth-provider/pkg/client"
	"github.com/obot-platform/providers/keycloak-auth-provider/pkg/config"
	"github.com/obot-platform/providers/keycloak-auth-provider/pkg/profile"
)

func TestGroupsToInfosPrefixesIDs(t *testing.T) {
	t.Parallel()

	got := groupsToInfos([]client.Group{{Path: "/developers"}})
	if len(got) != 1 {
		t.Fatalf("got %d groups, want 1", len(got))
	}
	wantID := config.GroupIDPrefix + "developers"
	if got[0].ID != wantID {
		t.Fatalf("ID = %q, want %q", got[0].ID, wantID)
	}
	if got[0].Name != "/developers" {
		t.Fatalf("Name = %q, want /developers", got[0].Name)
	}
}

func TestToGroupInfosPrefixesIDs(t *testing.T) {
	t.Parallel()

	got := toGroupInfos([]string{"/ops/elk"})
	if len(got) != 1 {
		t.Fatalf("got %d groups, want 1", len(got))
	}
	wantID := config.GroupIDPrefix + "ops/elk"
	if got[0].ID != wantID {
		t.Fatalf("ID = %q, want %q", got[0].ID, wantID)
	}
}

func TestProfileGroupInfosPrefixesIDs(t *testing.T) {
	t.Parallel()

	p := &profile.KeycloakProfile{FullGroupPath: []string{"/developers"}}
	got := p.GroupInfos()
	if len(got) != 1 {
		t.Fatalf("got %d groups, want 1", len(got))
	}
	wantID := config.GroupIDPrefix + "developers"
	if got[0].ID != wantID {
		t.Fatalf("ID = %q, want %q", got[0].ID, wantID)
	}
}

func TestWriteAuthGroupsPageIsHostEnvelope(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	h := &Handlers{}
	h.writeAuthGroupsPage(rec, groupsToInfos([]client.Group{{Path: "/developers"}}))

	if rec.Code != http.StatusOK && rec.Code != 0 {
		t.Fatalf("status = %d", rec.Code)
	}

	var page authGroupsPage
	if err := json.NewDecoder(rec.Body).Decode(&page); err != nil {
		t.Fatalf("decode page: %v", err)
	}
	if page.NextCursor != "" {
		t.Fatalf("nextCursor = %q, want empty", page.NextCursor)
	}
	if len(page.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(page.Items))
	}
	if page.Items[0].ID != config.GroupIDPrefix+"developers" {
		t.Fatalf("items[0].id = %q", page.Items[0].ID)
	}
}

func TestWriteAuthGroupsPageEmptyItemsIsArray(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	h := &Handlers{}
	h.writeAuthGroupsPage(rec, nil)

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(rec.Body).Decode(&raw); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if string(raw["items"]) != "[]" {
		t.Fatalf("items = %s, want []", raw["items"])
	}
}
