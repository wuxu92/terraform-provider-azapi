package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLatestStableVersion(t *testing.T) {
	dir := t.TempDir()
	for _, v := range []string{"2024-01-01", "2025-06-01", "2025-01-01", "2025-08-01-preview", "2020-08-01-preview"} {
		if err := os.Mkdir(filepath.Join(dir, v), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// A stray non-directory entry must be ignored.
	if err := os.WriteFile(filepath.Join(dir, "index.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LatestStableVersion(dir)
	if err != nil {
		t.Fatal(err)
	}
	// 2025-08-01-preview is newer but must be skipped; 2025-06-01 is latest stable.
	if got != "2025-06-01" {
		t.Errorf("LatestStableVersion = %q, want 2025-06-01", got)
	}
}

func TestLatestStableVersionNoStable(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "2020-08-01-preview"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := LatestStableVersion(dir); err == nil {
		t.Error("expected an error when only preview versions exist")
	}
}

// latestStorageDefs loads and parses the storage types.json for the latest stable
// API version, returning the parsed defs and that version. Generator fixture tests
// resolve through this (never a hardcoded version) so they always exercise the
// version the provider actually ships.
func latestStorageDefs(t *testing.T) ([]*ResourceDefinition, string) {
	t.Helper()
	dir := filepath.Join("..", "..", "azure", "generated", "storage", "microsoft.storage")
	ver, err := LatestStableVersion(dir)
	if err != nil {
		t.Skipf("no stable storage version: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, ver, "types.json"))
	if err != nil {
		t.Skipf("types.json not found: %v", err)
	}
	defs, err := ParseTypesJSON(data)
	if err != nil {
		t.Fatalf("ParseTypesJSON: %v", err)
	}
	return defs, ver
}
