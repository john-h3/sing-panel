package services

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"sing-panel/models"
)

func newTagTestService(t *testing.T) *SingBoxConfigService {
	t.Helper()
	db, err := NewDatabase(filepath.Join(t.TempDir(), "sing-panel.db"))
	if err != nil {
		t.Fatalf("create database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	configService := NewConfigService(db)
	return NewSingBoxConfigService(db, configService)
}

func requireDuplicateTag(t *testing.T, err error, resource, tag string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected duplicate %s tag error", resource)
	}
	var duplicate *DuplicateTagError
	if !errors.As(err, &duplicate) {
		t.Fatalf("error = %v, want DuplicateTagError", err)
	}
	if duplicate.Resource != resource || duplicate.Tag != tag {
		t.Fatalf("duplicate error = %#v, want resource=%q tag=%q", duplicate, resource, tag)
	}
}

func TestOutboundTagMustBeUniqueAndSelfUpdateIsAllowed(t *testing.T) {
	service := newTagTestService(t)
	first, err := service.AddOutbound(models.Outbound{Type: "direct", Tag: "proxy"})
	if err != nil {
		t.Fatalf("add first outbound: %v", err)
	}

	_, err = service.AddOutbound(models.Outbound{Type: "block", Tag: " proxy "})
	requireDuplicateTag(t, err, "outbound", "proxy")

	updated, err := service.UpdateOutbound(models.Outbound{ID: first.ID, Type: "direct", Tag: " proxy "})
	if err != nil {
		t.Fatalf("update outbound with its own tag: %v", err)
	}
	if updated.Tag != "proxy" {
		t.Fatalf("updated tag = %q, want normalized proxy", updated.Tag)
	}

	_, err = service.AddOutbound(models.Outbound{Type: "direct", Tag: "   "})
	if err == nil || !strings.HasSuffix(err.Error(), " tag is required") {
		t.Fatalf("blank tag error = %v, want required-tag error", err)
	}
}

func TestRulesetBatchTagsMustBeUnique(t *testing.T) {
	service := newTagTestService(t)
	_, err := service.AddRulesets([]models.Ruleset{
		{Type: "remote", Tag: "geo-a", Options: map[string]interface{}{"url": "https://example.com/a"}},
		{Type: "remote", Tag: "geo-a", Options: map[string]interface{}{"url": "https://example.com/b"}},
	})
	requireDuplicateTag(t, err, "ruleset", "geo-a")
}

func TestDNSServerTagsMustBeUnique(t *testing.T) {
	service := newTagTestService(t)
	_, err := service.GetDNS()
	if err != nil {
		t.Fatalf("get DNS: %v", err)
	}

	err = service.UpdateDNS(models.DNSConfig{Servers: []models.DNSServer{
		{Type: "local", Tag: "local"},
		{Type: "local", Tag: " local "},
	}})
	requireDuplicateTag(t, err, "dns server", "local")

	err = service.UpdateDNS(models.DNSConfig{Servers: []models.DNSServer{{Type: "local", Tag: " "}}})
	if err == nil || !strings.HasSuffix(err.Error(), " tag is required") {
		t.Fatalf("blank DNS tag error = %v, want required-tag error", err)
	}
}
