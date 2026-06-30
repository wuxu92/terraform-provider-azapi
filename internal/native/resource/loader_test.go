package resource

import "testing"

func TestLoadStorageBody(t *testing.T) {
	body, err := loadBody("Microsoft.Storage/storageAccounts", "2025-06-01")
	if err != nil {
		t.Fatalf("loadBody: %v", err)
	}
	if body == nil || body.Properties["sku"] == nil {
		t.Fatal("storage body missing sku")
	}
	// Cached on second call.
	body2, _ := loadBody("Microsoft.Storage/storageAccounts", "2025-06-01")
	if body2 != body {
		t.Error("loadBody not cached")
	}
}
