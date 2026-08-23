package services

import (
	"testing"

	"sing-panel/models"
)

func TestDisableCustomizedOutbounds(t *testing.T) {
	config := models.SingBoxConfig{
		Outbounds: []models.Outbound{
			{ID: "fallback", Type: "fallback", Tag: "fallback-out", Enabled: true},
			{ID: "loadbalance", Type: "loadbalance", Tag: "loadbalance-out", Enabled: true},
			{ID: "direct", Type: "direct", Tag: "direct-out", Enabled: true},
			{ID: "fallback-disabled", Type: "fallback", Tag: "old-fallback", Enabled: false},
		},
		RouteConfig: &models.RouteConfig{Final: "fallback-out"},
		RouteRules: []models.RouteRule{
			{ID: "rule-1", Outbound: "fallback-out"},
			{ID: "rule-2", Outbound: "direct-out"},
		},
	}

	disableCustomizedOutbounds(&config)

	if config.Outbounds[0].Enabled || config.Outbounds[1].Enabled || config.Outbounds[3].Enabled {
		t.Fatal("customized outbounds should be disabled")
	}
	if !config.Outbounds[2].Enabled {
		t.Fatal("standard outbound should remain enabled")
	}
	if config.RouteConfig.Final != "" {
		t.Fatalf("route final = %q, want empty", config.RouteConfig.Final)
	}
	if config.RouteRules[0].Outbound != "" {
		t.Fatalf("customized route outbound = %q, want empty", config.RouteRules[0].Outbound)
	}
	if config.RouteRules[1].Outbound != "direct-out" {
		t.Fatalf("standard route outbound = %q, want direct-out", config.RouteRules[1].Outbound)
	}
}

func TestBuildOutboundsFiltersCustomizedTypes(t *testing.T) {
	service := &SingBoxConfigService{}
	outbounds := []models.Outbound{
		{Type: "fallback", Tag: "fallback-out", Enabled: true, Options: map[string]interface{}{"outbounds": []string{"direct-out"}}},
		{Type: "direct", Tag: "direct-out", Enabled: true},
	}

	if got := service.buildOutbounds(outbounds, false); len(got) != 1 || got[0]["tag"] != "direct-out" {
		t.Fatalf("disabled customized export = %#v, want only direct-out", got)
	}
	if got := service.buildOutbounds(outbounds, true); len(got) != 2 {
		t.Fatalf("enabled customized export = %#v, want two outbounds", got)
	}
}

func TestCustomizedOutboundCannotBeEnabledWhenFeatureIsDisabled(t *testing.T) {
	config := models.AppConfig{}
	if customizedFeaturesEnabled(config) {
		t.Fatal("customized features should be disabled by default")
	}

	for _, outboundType := range []string{"fallback", "loadbalance"} {
		outbound := models.Outbound{Type: outboundType, Enabled: true}
		if isCustomizedOutboundType(outbound.Type) && !customizedFeaturesEnabled(config) {
			outbound.Enabled = false
		}
		if outbound.Enabled {
			t.Fatalf("%s should not be enabled while customized features are disabled", outboundType)
		}
	}
}
