package generator

import (
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/azure"
	"github.com/Azure/terraform-provider-azapi/internal/native/typegraph"
)

// latestStableDefs loads and parses the bicep types.json for armType's latest
// stable API version, resolved through the azure schema loader (never a hardcoded
// version) from the same embedded source the provider runtime loads. Generator
// fixture tests resolve through this so they always exercise the version the
// provider ships. Returns the parsed defs and the resolved version.
func latestStableDefs(t *testing.T, armType string) ([]*typegraph.ResourceDefinition, string) {
	t.Helper()
	version, err := azure.GetLatestStableApiVersion(armType)
	if err != nil {
		t.Skipf("no stable version for %s: %v", armType, err)
	}
	location, err := azure.GetResourceTypeLocation(armType, version)
	if err != nil {
		t.Skipf("no types.json for %s: %v", armType, err)
	}
	data, err := azure.StaticFiles.ReadFile("generated/" + location)
	if err != nil {
		t.Skipf("types.json not found: %v", err)
	}
	defs, err := typegraph.ParseTypesJSON(data)
	if err != nil {
		t.Fatalf("ParseTypesJSON: %v", err)
	}
	return defs, version
}

// latestStorageDefs returns the storage account defs for the latest stable
// API version.
func latestStorageDefs(t *testing.T) ([]*typegraph.ResourceDefinition, string) {
	t.Helper()
	return latestStableDefs(t, "Microsoft.Storage/storageAccounts")
}

// latestResourceGroupDefs returns the resource group defs for the latest stable
// API version, mirroring latestStorageDefs for the resources service.
func latestResourceGroupDefs(t *testing.T) ([]*typegraph.ResourceDefinition, string) {
	t.Helper()
	return latestStableDefs(t, "Microsoft.Resources/resourceGroups")
}

// latestWebServerFarmDefs returns the Web server farm defs for the latest stable API version.
func latestWebServerFarmDefs(t *testing.T) ([]*typegraph.ResourceDefinition, string) {
	t.Helper()
	return latestStableDefs(t, "Microsoft.Web/serverfarms")
}

// latestWebSiteDefs returns the Web site defs for the latest stable API version.
func latestWebSiteDefs(t *testing.T) ([]*typegraph.ResourceDefinition, string) {
	t.Helper()
	return latestStableDefs(t, "Microsoft.Web/sites")
}
