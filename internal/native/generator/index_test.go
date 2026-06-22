package generator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// writeTestIndex writes a temp index.json from a "<ARMType>@<version>" -> $ref map
// and returns it loaded.
func writeTestIndex(t *testing.T, resources map[string]string) *Index {
	t.Helper()
	entries := make(map[string]map[string]string, len(resources))
	for tag, ref := range resources {
		entries[tag] = map[string]string{"$ref": ref}
	}
	blob, err := json.Marshal(map[string]any{"resources": entries})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "index.json")
	if err := os.WriteFile(path, blob, 0o644); err != nil {
		t.Fatal(err)
	}
	idx, err := LoadIndex(path)
	if err != nil {
		t.Fatal(err)
	}
	return idx
}

func TestIndexLatestStableVersion(t *testing.T) {
	idx := writeTestIndex(t, map[string]string{
		"Microsoft.Storage/storageAccounts@2024-01-01":         "storage/microsoft.storage/2024-01-01/types.json#/0",
		"Microsoft.Storage/storageAccounts@2025-01-01":         "storage/microsoft.storage/2025-01-01/types.json#/0",
		"Microsoft.Storage/storageAccounts@2025-06-01":         "storage/microsoft.storage/2025-06-01/types.json#/0",
		"Microsoft.Storage/storageAccounts@2025-08-01-preview": "storage/microsoft.storage/2025-08-01-preview/types.json#/0",
		// A child type with a NEWER version must not leak into the parent lookup.
		"Microsoft.Storage/storageAccounts/blobServices@2026-01-01": "storage/microsoft.storage/2026-01-01/types.json#/3",
	})
	got, err := idx.LatestStableVersion("Microsoft.Storage/storageAccounts")
	if err != nil {
		t.Fatal(err)
	}
	if got != "2025-06-01" {
		t.Errorf("LatestStableVersion = %q, want 2025-06-01 (newest non-preview, child ignored)", got)
	}
}

func TestIndexLatestStableVersionNoStable(t *testing.T) {
	idx := writeTestIndex(t, map[string]string{
		"Microsoft.Foo/bars@2020-08-01-preview": "foo/microsoft.foo/2020-08-01-preview/types.json#/0",
	})
	if _, err := idx.LatestStableVersion("Microsoft.Foo/bars"); err == nil {
		t.Error("expected an error when only preview versions exist")
	}
}

func TestIndexTypesPath(t *testing.T) {
	idx := writeTestIndex(t, map[string]string{
		"Microsoft.Storage/storageAccounts@2025-06-01": "storage/microsoft.storage/2025-06-01/types.json#/29",
	})
	got, err := idx.TypesPath("Microsoft.Storage/storageAccounts@2025-06-01")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(idx.dir, "storage", "microsoft.storage", "2025-06-01", "types.json")
	if got != want {
		t.Errorf("TypesPath = %q, want %q (fragment stripped, joined to index dir)", got, want)
	}
	if _, err := idx.TypesPath("Microsoft.Storage/storageAccounts@1999-01-01"); err == nil {
		t.Error("expected an error for a tag absent from the index")
	}
}

// testIndexPath locates the real bicep-types manifest from the generator package.
func testIndexPath() string {
	return filepath.Join("..", "..", "azure", "generated", "index.json")
}

// latestStorageDefs loads and parses the storage account types.json for the
// latest stable API version (resolved through the manifest, never a hardcoded
// version), returning the parsed defs and that version. Generator fixture tests
// resolve through this so they always exercise the version the provider ships.
func latestStorageDefs(t *testing.T) ([]*ResourceDefinition, string) {
	t.Helper()
	idx, err := LoadIndex(testIndexPath())
	if err != nil {
		t.Skipf("no index.json: %v", err)
	}
	tag, typesPath, err := idx.ResolveLatestStable("Microsoft.Storage/storageAccounts")
	if err != nil {
		t.Skipf("no stable storage version: %v", err)
	}
	data, err := os.ReadFile(typesPath)
	if err != nil {
		t.Skipf("types.json not found: %v", err)
	}
	defs, err := ParseTypesJSON(data)
	if err != nil {
		t.Fatalf("ParseTypesJSON: %v", err)
	}
	ver := tag[len("Microsoft.Storage/storageAccounts@"):]
	return defs, ver
}

// latestResourceGroupDefs loads and parses the resources types.json for the latest
// stable resourceGroups API version (resolved through the manifest), mirroring
// latestStorageDefs for the resources service.
func latestResourceGroupDefs(t *testing.T) ([]*ResourceDefinition, string) {
	t.Helper()
	idx, err := LoadIndex(testIndexPath())
	if err != nil {
		t.Skipf("no index.json: %v", err)
	}
	tag, typesPath, err := idx.ResolveLatestStable("Microsoft.Resources/resourceGroups")
	if err != nil {
		t.Skipf("no stable resource group version: %v", err)
	}
	data, err := os.ReadFile(typesPath)
	if err != nil {
		t.Skipf("types.json not found: %v", err)
	}
	defs, err := ParseTypesJSON(data)
	if err != nil {
		t.Fatalf("ParseTypesJSON: %v", err)
	}
	ver := tag[len("Microsoft.Resources/resourceGroups@"):]
	return defs, ver
}
