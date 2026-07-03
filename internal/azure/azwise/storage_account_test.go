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

		// Tier-agnostic: the zone rule matches the replication suffix, so Premium,
		// PremiumV2 and StandardV2 cross-zone changes are covered too — not just Standard.
		{"Premium LRS to ZRS", "Premium_LRS", "Premium_ZRS", true},
		{"Premium ZRS to LRS", "Premium_ZRS", "Premium_LRS", true},
		{"PremiumV2 LRS to ZRS", "PremiumV2_LRS", "PremiumV2_ZRS", true},
		{"StandardV2 GRS to GZRS", "StandardV2_GRS", "StandardV2_GZRS", true},
		{"Premium LRS unchanged", "Premium_LRS", "Premium_LRS", false},
		// Tier change alone (same replication suffix) is not a zone migration; the
		// unconditional sku.tier ForceNew handles that path and is not exercised here.
		{"tier change same replication", "Standard_LRS", "Premium_LRS", false},
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

	// Verify base ForceNew rules still fire through the override
	t.Run("base rule isHnsEnabled", func(t *testing.T) {
		oldBody := map[string]interface{}{"properties": map[string]interface{}{"isHnsEnabled": false}}
		newBody := map[string]interface{}{"properties": map[string]interface{}{"isHnsEnabled": true}}
		if !k.CheckForceNew(oldBody, newBody) {
			t.Error("expected ForceNew when isHnsEnabled changes")
		}
	})
	t.Run("base rule dnsEndpointType", func(t *testing.T) {
		oldBody := map[string]interface{}{"properties": map[string]interface{}{"dnsEndpointType": "Standard"}}
		newBody := map[string]interface{}{"properties": map[string]interface{}{"dnsEndpointType": "AzureDnsZone"}}
		if !k.CheckForceNew(oldBody, newBody) {
			t.Error("expected ForceNew when dnsEndpointType changes")
		}
	})
	t.Run("base rule no change", func(t *testing.T) {
		oldBody := map[string]interface{}{"properties": map[string]interface{}{"isHnsEnabled": true}}
		newBody := map[string]interface{}{"properties": map[string]interface{}{"isHnsEnabled": true}}
		if k.CheckForceNew(oldBody, newBody) {
			t.Error("expected no ForceNew when isHnsEnabled unchanged")
		}
	})
}

// ---------------------------------------------------------------------------
// Storage Account: account_kind migration ForceNew
// ---------------------------------------------------------------------------

func TestStorageCheckForceNewAccountKind(t *testing.T) {
	k := NewStorageAccount()

	tests := []struct {
		name    string
		oldKind string
		newKind string
		expect  bool
	}{
		// Storage -> StorageV2 is the one supported in-place migration.
		{"Storage to StorageV2 in place", "Storage", "StorageV2", false},
		{"any to StorageV2 in place", "BlobStorage", "StorageV2", false},
		{"Storage to any in place", "Storage", "BlockBlobStorage", false},

		// Every other kind change forces replacement.
		{"StorageV2 to BlobStorage replaces", "StorageV2", "BlobStorage", true},
		{"BlobStorage to FileStorage replaces", "BlobStorage", "FileStorage", true},

		// No change / create.
		{"unchanged", "StorageV2", "StorageV2", false},
		{"create (no old kind)", "", "StorageV2", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			oldBody := map[string]interface{}{}
			newBody := map[string]interface{}{}
			if tc.oldKind != "" {
				oldBody["kind"] = tc.oldKind
			}
			if tc.newKind != "" {
				newBody["kind"] = tc.newKind
			}
			if got := k.CheckForceNew(oldBody, newBody); got != tc.expect {
				t.Errorf("CheckForceNew(kind %q → %q) = %v, want %v", tc.oldKind, tc.newKind, got, tc.expect)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Storage Account: large_file_shares_state disablement ForceNew
// ---------------------------------------------------------------------------

func TestStorageCheckForceNewLargeFileShare(t *testing.T) {
	k := NewStorageAccount()

	tests := []struct {
		name   string
		oldLFS string
		newLFS string
		expect bool
	}{
		// Enabled cannot be turned off in place.
		{"enabled to disabled replaces", "Enabled", "Disabled", true},
		{"enabled to absent replaces", "Enabled", "", true},

		// Enabling (or no change) is an in-place update.
		{"disabled to enabled in place", "Disabled", "Enabled", false},
		{"absent to enabled in place", "", "Enabled", false},
		{"enabled unchanged", "Enabled", "Enabled", false},
		{"disabled unchanged", "Disabled", "Disabled", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			oldBody := map[string]interface{}{}
			newBody := map[string]interface{}{}
			if tc.oldLFS != "" {
				oldBody["properties"] = map[string]interface{}{"largeFileSharesState": tc.oldLFS}
			}
			if tc.newLFS != "" {
				newBody["properties"] = map[string]interface{}{"largeFileSharesState": tc.newLFS}
			}
			if got := k.CheckForceNew(oldBody, newBody); got != tc.expect {
				t.Errorf("CheckForceNew(largeFileSharesState %q → %q) = %v, want %v", tc.oldLFS, tc.newLFS, got, tc.expect)
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
	if len(fields) != 6 {
		t.Fatalf("expected 6 sensitive fields, got %d", len(fields))
	}
	expected := map[string]bool{
		"properties.primaryAccessKey":              true,
		"properties.secondaryAccessKey":            true,
		"properties.primaryConnectionString":       true,
		"properties.secondaryConnectionString":     true,
		"properties.primaryBlobConnectionString":   true,
		"properties.secondaryBlobConnectionString": true,
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
			t.Errorf("  %s: %s", d.Summary(), d.Detail())
		}
	}

	// A sensitive field set in body without sensitive_body should be flagged.
	diags = k.Validate("", map[string]interface{}{
		"sku":  map[string]interface{}{"name": "Standard_LRS"},
		"kind": "StorageV2",
		"properties": map[string]interface{}{
			"primaryAccessKey": "secret",
		},
	}, false)
	if len(diags) != 1 {
		t.Errorf("expected 1 diagnostic (sensitive), got %d:", len(diags))
		for _, d := range diags {
			t.Errorf("  %s: %s", d.Summary(), d.Detail())
		}
	}
	if len(diags) == 1 && diags[0].Summary() != "Sensitive properties should use sensitive_body" {
		t.Errorf("unexpected summary: %s", diags[0].Summary())
	}

	// No sensitive field in body → no sensitive diagnostic.
	diags = k.Validate("", map[string]interface{}{
		"sku":  map[string]interface{}{"name": "Standard_LRS"},
		"kind": "StorageV2",
	}, false)
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics when no sensitive field in body, got %d:", len(diags))
		for _, d := range diags {
			t.Errorf("  %s: %s", d.Summary(), d.Detail())
		}
	}

	// Missing required fields should produce required-fields diagnostic
	diags = k.Validate("", map[string]interface{}{}, true)
	if len(diags) != 1 {
		t.Errorf("expected 1 diagnostic (required), got %d:", len(diags))
		for _, d := range diags {
			t.Errorf("  %s: %s", d.Summary(), d.Detail())
		}
	}
	if len(diags) == 1 && diags[0].Summary() != "Missing required properties" {
		t.Errorf("unexpected summary: %s", diags[0].Summary())
	}
}

// ---------------------------------------------------------------------------
// Storage Account: computed fields coverage
// ---------------------------------------------------------------------------

func TestStorageComputedFields(t *testing.T) {
	k := NewStorageAccount()
	if len(k.ComputedFields) != 17 {
		t.Fatalf("expected 17 computed fields, got %d", len(k.ComputedFields))
	}
	// Spot-check critical entries
	want := map[string]bool{
		"properties.primaryEndpoints":                  true,
		"properties.secondaryEndpoints":                true,
		"properties.provisioningState":                 true,
		"properties.primaryLocation":                   true,
		"properties.privateEndpointConnections":        true,
		"properties.storageAccountSkuConversionStatus": true,
	}
	for _, f := range k.ComputedFields {
		delete(want, f)
	}
	for f := range want {
		t.Errorf("missing expected computed field: %s", f)
	}
}

func TestStorageForceNewRules(t *testing.T) {
	k := NewStorageAccount()
	if len(k.ForceNew) != 9 {
		t.Fatalf("expected 9 ForceNew rules, got %d", len(k.ForceNew))
	}
	want := map[string]bool{
		"sku.tier":                   true,
		"properties.isHnsEnabled":    true,
		"extendedLocation":           true,
		"properties.dnsEndpointType": true,
	}
	for _, r := range k.ForceNew {
		delete(want, r.PropertyPath)
	}
	for f := range want {
		t.Errorf("missing expected ForceNew rule: %s", f)
	}
}

func TestStorageDefaultValues(t *testing.T) {
	k := NewStorageAccount()
	if len(k.DefaultValues) != 18 {
		t.Fatalf("expected 18 default values, got %d", len(k.DefaultValues))
	}
	// Spot-check a few key defaults
	defaults := make(map[string]interface{})
	for _, dv := range k.DefaultValues {
		defaults[dv.PropertyPath] = dv.Value
	}
	if defaults["kind"] != "StorageV2" {
		t.Errorf("kind default: got %v, want StorageV2", defaults["kind"])
	}
	if defaults["properties.minimumTlsVersion"] != "TLS1_2" {
		t.Errorf("minimumTlsVersion default: got %v, want TLS1_2", defaults["properties.minimumTlsVersion"])
	}
	if defaults["properties.allowBlobPublicAccess"] != false {
		t.Errorf("allowBlobPublicAccess default: got %v, want false", defaults["properties.allowBlobPublicAccess"])
	}
}
