package azwise

import (
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Storage Account: interface compliance
// ---------------------------------------------------------------------------

func TestStorageAccountImplementsInterface(t *testing.T) {
	var _ ResourceKnowledge = NewStorageAccount()
}

// ---------------------------------------------------------------------------
// Storage Account: SKU zone migration ForceNew
// ---------------------------------------------------------------------------

func TestStorageCheckForceNew(t *testing.T) {
	k := NewStorageAccount()

	tests := []struct {
		name   string
		oldSku string
		newSku string
		expect bool
	}{
		// Same SKU — no replacement
		{"same LRS", "Standard_LRS", "Standard_LRS", false},
		{"same ZRS", "Standard_ZRS", "Standard_ZRS", false},

		// Non-zonal to zonal — replacement
		{"LRS to ZRS", "Standard_LRS", "Standard_ZRS", true},
		{"GRS to ZRS", "Standard_GRS", "Standard_ZRS", true},
		{"RAGRS to GZRS", "Standard_RAGRS", "Standard_GZRS", true},
		{"LRS to RAGZRS", "Standard_LRS", "Standard_RAGZRS", true},

		// Zonal to non-zonal — replacement
		{"ZRS to LRS", "Standard_ZRS", "Standard_LRS", true},
		{"GZRS to GRS", "Standard_GZRS", "Standard_GRS", true},
		{"RAGZRS to RAGRS", "Standard_RAGZRS", "Standard_RAGRS", true},

		// Within same zone category — no replacement
		{"LRS to GRS", "Standard_LRS", "Standard_GRS", false},
		{"GRS to RAGRS", "Standard_GRS", "Standard_RAGRS", false},
		{"ZRS to GZRS", "Standard_ZRS", "Standard_GZRS", false},
		{"GZRS to RAGZRS", "Standard_GZRS", "Standard_RAGZRS", false},

		// Missing SKU — no replacement
		{"missing old", "", "Standard_LRS", false},
		{"missing new", "Standard_LRS", "", false},
		{"both missing", "", "", false},

		// Case insensitivity
		{"case insensitive", "standard_lrs", "STANDARD_ZRS", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			oldBody := map[string]interface{}{}
			newBody := map[string]interface{}{}
			if tc.oldSku != "" {
				oldBody["sku"] = map[string]interface{}{"name": tc.oldSku}
			}
			if tc.newSku != "" {
				newBody["sku"] = map[string]interface{}{"name": tc.newSku}
			}

			got := k.CheckForceNew(oldBody, newBody)
			if got != tc.expect {
				t.Errorf("CheckForceNew(%s → %s) = %v, want %v", tc.oldSku, tc.newSku, got, tc.expect)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Storage Account: naming validation
// ---------------------------------------------------------------------------

func TestStorageValidateName(t *testing.T) {
	k := NewStorageAccount()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "mystorageaccount1", false},
		{"min length", "abc", false},
		{"max length", "abcdefghijklmnopqrstuvwx", false}, // 24 chars
		{"too short", "ab", true},
		{"too long", "abcdefghijklmnopqrstuvwxy", true}, // 25 chars
		{"uppercase", "MyStorageAccount", true},
		{"hyphens", "my-storage-account", true},
		{"underscores", "my_storage", true},
		{"digits only", "123456789012", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := k.ValidateName(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateName(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Storage Account: body string property validation
// ---------------------------------------------------------------------------

func TestStorageValidatePropertiesStringRules(t *testing.T) {
	k := NewStorageAccount()

	tests := []struct {
		name      string
		body      map[string]interface{}
		wantCount int
	}{
		{
			"valid sku and kind",
			map[string]interface{}{
				"sku":  map[string]interface{}{"name": "Standard_LRS"},
				"kind": "StorageV2",
			},
			0,
		},
		{
			"valid sku case insensitive",
			map[string]interface{}{
				"sku": map[string]interface{}{"name": "standard_lrs"},
			},
			0,
		},
		{
			"invalid sku",
			map[string]interface{}{
				"sku": map[string]interface{}{"name": "Invalid_SKU"},
			},
			1,
		},
		{
			"invalid kind",
			map[string]interface{}{
				"kind": "InvalidKind",
			},
			1,
		},
		{
			"invalid access tier",
			map[string]interface{}{
				"properties": map[string]interface{}{"accessTier": "Warm"},
			},
			1,
		},
		{
			"valid access tier",
			map[string]interface{}{
				"properties": map[string]interface{}{"accessTier": "Hot"},
			},
			0,
		},
		{
			"multiple invalid",
			map[string]interface{}{
				"sku":  map[string]interface{}{"name": "Bad"},
				"kind": "Bad",
			},
			2,
		},
		{
			"absent properties",
			map[string]interface{}{},
			0,
		},
		{
			"all valid SKUs",
			map[string]interface{}{
				"sku": map[string]interface{}{"name": "Premium_ZRS"},
			},
			0,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			errs := k.ValidateProperties(tc.body)
			if len(errs) != tc.wantCount {
				t.Errorf("ValidateProperties() got %d errors, want %d: %v", len(errs), tc.wantCount, errs)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Storage Account: timeouts
// ---------------------------------------------------------------------------

func TestStorageTimeouts(t *testing.T) {
	k := NewStorageAccount()
	fallback := 30 * time.Minute

	if got := k.TimeoutDefault("create", fallback); got != 60*time.Minute {
		t.Errorf("create = %v, want 60m", got)
	}
	if got := k.TimeoutDefault("update", fallback); got != 60*time.Minute {
		t.Errorf("update = %v, want 60m", got)
	}
	if got := k.TimeoutDefault("delete", fallback); got != 60*time.Minute {
		t.Errorf("delete = %v, want 60m", got)
	}
	if got := k.TimeoutDefault("read", fallback); got != 5*time.Minute {
		t.Errorf("read = %v, want 5m", got)
	}
}

// ---------------------------------------------------------------------------
// Storage Account: sensitive fields
// ---------------------------------------------------------------------------

func TestStorageSensitiveFields(t *testing.T) {
	k := NewStorageAccount()
	fields := k.SensitiveFields
	if len(fields) != 2 {
		t.Fatalf("expected 2 sensitive fields, got %d", len(fields))
	}
	expected := map[string]bool{
		"properties.primaryAccessKey":   true,
		"properties.secondaryAccessKey": true,
	}
	for _, f := range fields {
		if !expected[f] {
			t.Errorf("unexpected sensitive field: %s", f)
		}
	}
}

// ---------------------------------------------------------------------------
// Storage Account: Validate (combined)
// ---------------------------------------------------------------------------

func TestStorageValidate(t *testing.T) {
	k := NewStorageAccount()

	// Valid configuration with sensitive_body
	diags := k.Validate("mystorage123", map[string]interface{}{
		"sku":  map[string]interface{}{"name": "Standard_LRS"},
		"kind": "StorageV2",
	}, true)
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics, got %d:", len(diags))
		for _, d := range diags {
			t.Errorf("  %s: %s", d.Summary, d.Detail)
		}
	}

	// Missing sensitive_body should produce sensitive field diagnostic
	diags = k.Validate("", map[string]interface{}{}, false)
	if len(diags) != 1 {
		t.Errorf("expected 1 diagnostic (sensitive), got %d:", len(diags))
		for _, d := range diags {
			t.Errorf("  %s: %s", d.Summary, d.Detail)
		}
	}
	if len(diags) == 1 && diags[0].Summary != "Sensitive properties should use sensitive_body" {
		t.Errorf("unexpected summary: %s", diags[0].Summary)
	}
}
