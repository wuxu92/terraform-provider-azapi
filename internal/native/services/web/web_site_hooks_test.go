package web

import "testing"

func TestWebSiteConfigurationID(t *testing.T) {
	siteID := "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Web/sites/site"
	if got := webSiteConfigurationID(siteID); got != siteID+"/config/web" {
		t.Fatalf("webSiteConfigurationID = %q, want config/web child id", got)
	}
}

func TestNormalizeWebSiteResponseDefaultsMissingAffinityPartitioning(t *testing.T) {
	site := map[string]interface{}{
		"properties": map[string]interface{}{},
	}
	config := map[string]interface{}{
		"properties": map[string]interface{}{
			"ftpsState": "Disabled",
		},
	}

	normalizeWebSiteResponse(site, config)

	props := site["properties"].(map[string]interface{})
	if props["clientAffinityPartitioningEnabled"] != false {
		t.Fatalf("clientAffinityPartitioningEnabled = %#v, want false when API omits it", props["clientAffinityPartitioningEnabled"])
	}
	if _, ok := props["siteConfig"]; !ok {
		t.Fatalf("siteConfig was not merged: %#v", props)
	}
}

func TestNormalizeWebSiteResponsePreservesReturnedAffinityPartitioning(t *testing.T) {
	site := map[string]interface{}{
		"properties": map[string]interface{}{
			"clientAffinityPartitioningEnabled": true,
		},
	}
	normalizeWebSiteResponse(site, map[string]interface{}{})

	props := site["properties"].(map[string]interface{})
	if props["clientAffinityPartitioningEnabled"] != true {
		t.Fatalf("clientAffinityPartitioningEnabled = %#v, want returned true preserved", props["clientAffinityPartitioningEnabled"])
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

func TestMergeWebSiteConfigurationFiltersDefaultAccessRestrictions(t *testing.T) {
	site := map[string]interface{}{}
	config := map[string]interface{}{
		"properties": map[string]interface{}{
			"ipSecurityRestrictions": []interface{}{
				map[string]interface{}{
					"action":      "Allow",
					"description": "Allow all access",
					"ipAddress":   "Any",
					"name":        "Allow all",
					"priority":    int64(2147483647),
				},
				map[string]interface{}{
					"action":    "Deny",
					"ipAddress": "192.0.2.0/24",
					"name":      "deny-doc-range",
					"priority":  float64(100),
				},
			},
			"scmIpSecurityRestrictions": []interface{}{
				map[string]interface{}{
					"action":      "Allow",
					"description": "Allow all access",
					"ipAddress":   "Any",
					"name":        "Allow all",
					"priority":    float64(2147483647),
				},
			},
		},
	}

	mergeWebSiteConfiguration(site, config)

	siteConfig := site["properties"].(map[string]interface{})["siteConfig"].(map[string]interface{})
	restrictions := siteConfig["ipSecurityRestrictions"].([]interface{})
	if len(restrictions) != 1 || restrictions[0].(map[string]interface{})["name"] != "deny-doc-range" {
		t.Fatalf("ipSecurityRestrictions = %#v, want only configured deny rule", restrictions)
	}
	if _, ok := siteConfig["scmIpSecurityRestrictions"]; ok {
		t.Fatalf("scmIpSecurityRestrictions = %#v, want default allow-all rule removed", siteConfig["scmIpSecurityRestrictions"])
	}
}
