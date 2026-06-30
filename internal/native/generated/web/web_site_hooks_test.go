package web

import "testing"

func TestWebSiteConfigurationID(t *testing.T) {
	siteID := "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Web/sites/site"
	if got := webSiteConfigurationID(siteID); got != siteID+"/config/web" {
		t.Fatalf("webSiteConfigurationID = %q, want config/web child id", got)
	}
}

func TestMergeWebSiteConfiguration(t *testing.T) {
	site := map[string]interface{}{
		"properties": map[string]interface{}{
			"enabled": true,
		},
	}
	config := map[string]interface{}{
		"properties": map[string]interface{}{
			"ftpsState":        "Disabled",
			"minTlsVersion":    "1.2",
			"scmMinTlsVersion": "1.2",
		},
	}
	mergeWebSiteConfiguration(site, config)

	props := site["properties"].(map[string]interface{})
	if props["enabled"] != true {
		t.Fatalf("existing site properties were not preserved: %#v", props)
	}
	siteConfig := props["siteConfig"].(map[string]interface{})
	if siteConfig["ftpsState"] != "Disabled" || siteConfig["minTlsVersion"] != "1.2" || siteConfig["scmMinTlsVersion"] != "1.2" {
		t.Fatalf("siteConfig = %#v, want config/web properties merged", siteConfig)
	}
}
