package azwise

import (
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Key Vault: interface compliance
// ---------------------------------------------------------------------------

func TestKeyVaultImplementsInterface(t *testing.T) {
	var _ ResourceKnowledge = NewKeyVault()
}

// ---------------------------------------------------------------------------
// Key Vault: ForceNew
// ---------------------------------------------------------------------------

func TestKeyVaultCheckForceNew(t *testing.T) {
	k := NewKeyVault()

	tests := []struct {
		name   string
		old    map[string]interface{}
		new    map[string]interface{}
		expect bool
	}{
		{"name unchanged", map[string]interface{}{"name": "myvault"}, map[string]interface{}{"name": "myvault"}, false},
		{"name changed", map[string]interface{}{"name": "myvault"}, map[string]interface{}{"name": "other"}, true},
		{"name absent both", map[string]interface{}{}, map[string]interface{}{}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := k.CheckForceNew(tc.old, tc.new)
			if got != tc.expect {
				t.Errorf("CheckForceNew = %v, want %v", got, tc.expect)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Key Vault: naming validation
// ---------------------------------------------------------------------------

func TestKeyVaultValidateName(t *testing.T) {
	k := NewKeyVault()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "myvault", false},
		{"valid with hyphens", "my-vault-123", false},
		{"min length", "abc", false},
		{"max length 24", "abcdefghijklmnopqrstuvwx", false},
		{"too short", "ab", true},
		{"too long 25", "abcdefghijklmnopqrstuvwxy", true},
		{"starts with digit", "1vault", true},
		{"starts with hyphen", "-vault", true},
		{"ends with hyphen", "vault-", true},
		{"consecutive hyphens", "hello--world", true},
		{"underscores", "my_vault", true},
		{"uppercase valid", "MyVault", false},
		{"only letters", "abcdef", false},
		{"digits only", "20202020", true},
		{"special chars", "ABC123!@", true},
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
// Key Vault: property validation — all string rules
// ---------------------------------------------------------------------------

func TestKeyVaultValidateProperties(t *testing.T) {
	k := NewKeyVault()

	tests := []struct {
		name      string
		body      map[string]interface{}
		wantCount int
	}{
		// ── sku.family ──
		{
			"valid sku family A",
			map[string]interface{}{"properties": map[string]interface{}{
				"sku": map[string]interface{}{"family": "A"},
			}},
			0,
		},
		{
			"invalid sku family B",
			map[string]interface{}{"properties": map[string]interface{}{
				"sku": map[string]interface{}{"family": "B"},
			}},
			1,
		},
		// ── sku.name ──
		{
			"valid sku standard",
			map[string]interface{}{"properties": map[string]interface{}{
				"sku": map[string]interface{}{"name": "standard"},
			}},
			0,
		},
		{
			"valid sku premium",
			map[string]interface{}{"properties": map[string]interface{}{
				"sku": map[string]interface{}{"name": "premium"},
			}},
			0,
		},
		{
			"valid sku case insensitive",
			map[string]interface{}{"properties": map[string]interface{}{
				"sku": map[string]interface{}{"name": "Standard"},
			}},
			0,
		},
		{
			"invalid sku basic",
			map[string]interface{}{"properties": map[string]interface{}{
				"sku": map[string]interface{}{"name": "basic"},
			}},
			1,
		},
		// ── createMode ──
		{
			"valid createMode default",
			map[string]interface{}{"properties": map[string]interface{}{
				"createMode": "default",
			}},
			0,
		},
		{
			"valid createMode recover",
			map[string]interface{}{"properties": map[string]interface{}{
				"createMode": "recover",
			}},
			0,
		},
		{
			"invalid createMode",
			map[string]interface{}{"properties": map[string]interface{}{
				"createMode": "purge",
			}},
			1,
		},
		// ── publicNetworkAccess ──
		{
			"valid publicNetworkAccess Enabled",
			map[string]interface{}{"properties": map[string]interface{}{
				"publicNetworkAccess": "Enabled",
			}},
			0,
		},
		{
			"valid publicNetworkAccess Disabled",
			map[string]interface{}{"properties": map[string]interface{}{
				"publicNetworkAccess": "Disabled",
			}},
			0,
		},
		{
			"invalid publicNetworkAccess",
			map[string]interface{}{"properties": map[string]interface{}{
				"publicNetworkAccess": "true",
			}},
			1,
		},
		// ── networkAcls.defaultAction ──
		{
			"valid defaultAction Allow",
			map[string]interface{}{"properties": map[string]interface{}{
				"networkAcls": map[string]interface{}{"defaultAction": "Allow"},
			}},
			0,
		},
		{
			"valid defaultAction Deny",
			map[string]interface{}{"properties": map[string]interface{}{
				"networkAcls": map[string]interface{}{"defaultAction": "Deny"},
			}},
			0,
		},
		{
			"invalid defaultAction",
			map[string]interface{}{"properties": map[string]interface{}{
				"networkAcls": map[string]interface{}{"defaultAction": "Block"},
			}},
			1,
		},
		// ── networkAcls.bypass ──
		{
			"valid bypass AzureServices",
			map[string]interface{}{"properties": map[string]interface{}{
				"networkAcls": map[string]interface{}{"bypass": "AzureServices"},
			}},
			0,
		},
		{
			"valid bypass None",
			map[string]interface{}{"properties": map[string]interface{}{
				"networkAcls": map[string]interface{}{"bypass": "None"},
			}},
			0,
		},
		{
			"invalid bypass",
			map[string]interface{}{"properties": map[string]interface{}{
				"networkAcls": map[string]interface{}{"bypass": "All"},
			}},
			1,
		},
		// ── softDeleteRetentionInDays (integer) ──
		{
			"valid retention 30",
			map[string]interface{}{"properties": map[string]interface{}{
				"softDeleteRetentionInDays": float64(30),
			}},
			0,
		},
		{
			"min retention boundary 7",
			map[string]interface{}{"properties": map[string]interface{}{
				"softDeleteRetentionInDays": float64(7),
			}},
			0,
		},
		{
			"max retention boundary 90",
			map[string]interface{}{"properties": map[string]interface{}{
				"softDeleteRetentionInDays": float64(90),
			}},
			0,
		},
		{
			"below min retention",
			map[string]interface{}{"properties": map[string]interface{}{
				"softDeleteRetentionInDays": float64(3),
			}},
			1,
		},
		{
			"above max retention",
			map[string]interface{}{"properties": map[string]interface{}{
				"softDeleteRetentionInDays": float64(100),
			}},
			1,
		},
		{
			"fractional retention rejected",
			map[string]interface{}{"properties": map[string]interface{}{
				"softDeleteRetentionInDays": float64(30.5),
			}},
			1,
		},
		// ── combined errors ──
		{
			"multiple invalid properties",
			map[string]interface{}{"properties": map[string]interface{}{
				"sku":                       map[string]interface{}{"name": "basic", "family": "B"},
				"createMode":                "purge",
				"softDeleteRetentionInDays": float64(3),
			}},
			4, // sku.name + sku.family + createMode + retention
		},
		// ── absent ──
		{
			"absent properties",
			map[string]interface{}{},
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
// Key Vault: ip_rules validation
// ---------------------------------------------------------------------------

func TestKeyVaultValidateIPRules(t *testing.T) {
	k := NewKeyVault()

	tests := []struct {
		name      string
		body      map[string]interface{}
		wantCount int
	}{
		{
			"valid IPv4 address",
			map[string]interface{}{"properties": map[string]interface{}{
				"networkAcls": map[string]interface{}{
					"ipRules": []interface{}{
						map[string]interface{}{"value": "10.0.0.1"},
					},
				},
			}},
			0,
		},
		{
			"valid CIDR block",
			map[string]interface{}{"properties": map[string]interface{}{
				"networkAcls": map[string]interface{}{
					"ipRules": []interface{}{
						map[string]interface{}{"value": "10.0.0.0/24"},
					},
				},
			}},
			0,
		},
		{
			"valid /32 single host",
			map[string]interface{}{"properties": map[string]interface{}{
				"networkAcls": map[string]interface{}{
					"ipRules": []interface{}{
						map[string]interface{}{"value": "192.168.1.1/32"},
					},
				},
			}},
			0,
		},
		{
			"multiple valid rules",
			map[string]interface{}{"properties": map[string]interface{}{
				"networkAcls": map[string]interface{}{
					"ipRules": []interface{}{
						map[string]interface{}{"value": "10.0.0.1"},
						map[string]interface{}{"value": "172.16.0.0/12"},
						map[string]interface{}{"value": "192.168.0.0/16"},
					},
				},
			}},
			0,
		},
		{
			"invalid ip - not an IP",
			map[string]interface{}{"properties": map[string]interface{}{
				"networkAcls": map[string]interface{}{
					"ipRules": []interface{}{
						map[string]interface{}{"value": "not-an-ip"},
					},
				},
			}},
			1,
		},
		{
			"invalid CIDR prefix /33",
			map[string]interface{}{"properties": map[string]interface{}{
				"networkAcls": map[string]interface{}{
					"ipRules": []interface{}{
						map[string]interface{}{"value": "10.0.0.0/33"},
					},
				},
			}},
			1,
		},
		{
			"mixed valid and invalid",
			map[string]interface{}{"properties": map[string]interface{}{
				"networkAcls": map[string]interface{}{
					"ipRules": []interface{}{
						map[string]interface{}{"value": "10.0.0.1"},
						map[string]interface{}{"value": "bad-value"},
					},
				},
			}},
			1,
		},
		{
			"empty ipRules array",
			map[string]interface{}{"properties": map[string]interface{}{
				"networkAcls": map[string]interface{}{
					"ipRules": []interface{}{},
				},
			}},
			0,
		},
		{
			"no ipRules field",
			map[string]interface{}{"properties": map[string]interface{}{
				"networkAcls": map[string]interface{}{},
			}},
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
// Key Vault: UUID field validation
// ---------------------------------------------------------------------------

func TestKeyVaultValidateUUIDs(t *testing.T) {
	k := NewKeyVault()

	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	invalidUUID := "not-a-uuid"

	tests := []struct {
		name      string
		body      map[string]interface{}
		wantCount int
	}{
		{
			"valid tenantId",
			map[string]interface{}{"properties": map[string]interface{}{
				"tenantId": validUUID,
			}},
			0,
		},
		{
			"invalid tenantId",
			map[string]interface{}{"properties": map[string]interface{}{
				"tenantId": invalidUUID,
			}},
			1,
		},
		{
			"valid accessPolicy UUIDs",
			map[string]interface{}{"properties": map[string]interface{}{
				"accessPolicies": []interface{}{
					map[string]interface{}{
						"tenantId": validUUID,
						"objectId": validUUID,
					},
				},
			}},
			0,
		},
		{
			"invalid accessPolicy tenantId",
			map[string]interface{}{"properties": map[string]interface{}{
				"accessPolicies": []interface{}{
					map[string]interface{}{
						"tenantId": invalidUUID,
						"objectId": validUUID,
					},
				},
			}},
			1,
		},
		{
			"invalid objectId in second policy",
			map[string]interface{}{"properties": map[string]interface{}{
				"accessPolicies": []interface{}{
					map[string]interface{}{
						"tenantId": validUUID,
						"objectId": validUUID,
					},
					map[string]interface{}{
						"tenantId": validUUID,
						"objectId": invalidUUID,
					},
				},
			}},
			1,
		},
		{
			"uppercase UUID accepted",
			map[string]interface{}{"properties": map[string]interface{}{
				"tenantId": "550E8400-E29B-41D4-A716-446655440000",
			}},
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
// Key Vault: access policy permission validation
// ---------------------------------------------------------------------------

func TestKeyVaultValidatePermissions(t *testing.T) {
	k := NewKeyVault()

	tests := []struct {
		name      string
		body      map[string]interface{}
		wantCount int
	}{
		{
			"valid key permissions",
			map[string]interface{}{"properties": map[string]interface{}{
				"accessPolicies": []interface{}{
					map[string]interface{}{
						"permissions": map[string]interface{}{
							"keys": []interface{}{"Get", "List", "Create", "Delete"},
						},
					},
				},
			}},
			0,
		},
		{
			"valid certificate permissions",
			map[string]interface{}{"properties": map[string]interface{}{
				"accessPolicies": []interface{}{
					map[string]interface{}{
						"permissions": map[string]interface{}{
							"certificates": []interface{}{"Get", "List", "ManageContacts", "Purge"},
						},
					},
				},
			}},
			0,
		},
		{
			"valid secret permissions",
			map[string]interface{}{"properties": map[string]interface{}{
				"accessPolicies": []interface{}{
					map[string]interface{}{
						"permissions": map[string]interface{}{
							"secrets": []interface{}{"Get", "List", "Set", "Purge"},
						},
					},
				},
			}},
			0,
		},
		{
			"valid storage permissions",
			map[string]interface{}{"properties": map[string]interface{}{
				"accessPolicies": []interface{}{
					map[string]interface{}{
						"permissions": map[string]interface{}{
							"storage": []interface{}{"Get", "List", "Set", "DeleteSAS", "RegenerateKey"},
						},
					},
				},
			}},
			0,
		},
		{
			"invalid key permission",
			map[string]interface{}{"properties": map[string]interface{}{
				"accessPolicies": []interface{}{
					map[string]interface{}{
						"permissions": map[string]interface{}{
							"keys": []interface{}{"Get", "InvalidPerm"},
						},
					},
				},
			}},
			1,
		},
		{
			"invalid secret permission",
			map[string]interface{}{"properties": map[string]interface{}{
				"accessPolicies": []interface{}{
					map[string]interface{}{
						"permissions": map[string]interface{}{
							"secrets": []interface{}{"Read"},
						},
					},
				},
			}},
			1,
		},
		{
			"invalid certificate permission",
			map[string]interface{}{"properties": map[string]interface{}{
				"accessPolicies": []interface{}{
					map[string]interface{}{
						"permissions": map[string]interface{}{
							"certificates": []interface{}{"All"},
						},
					},
				},
			}},
			1,
		},
		{
			"invalid storage permission",
			map[string]interface{}{"properties": map[string]interface{}{
				"accessPolicies": []interface{}{
					map[string]interface{}{
						"permissions": map[string]interface{}{
							"storage": []interface{}{"ReadWrite"},
						},
					},
				},
			}},
			1,
		},
		{
			"multiple invalid across two policies",
			map[string]interface{}{"properties": map[string]interface{}{
				"accessPolicies": []interface{}{
					map[string]interface{}{
						"permissions": map[string]interface{}{
							"keys": []interface{}{"BadKey"},
						},
					},
					map[string]interface{}{
						"permissions": map[string]interface{}{
							"secrets": []interface{}{"BadSecret"},
						},
					},
				},
			}},
			2,
		},
		{
			"case-insensitive permission match",
			map[string]interface{}{"properties": map[string]interface{}{
				"accessPolicies": []interface{}{
					map[string]interface{}{
						"permissions": map[string]interface{}{
							"keys": []interface{}{"get", "LIST"},
						},
					},
				},
			}},
			0,
		},
		{
			"empty permissions array",
			map[string]interface{}{"properties": map[string]interface{}{
				"accessPolicies": []interface{}{
					map[string]interface{}{
						"permissions": map[string]interface{}{
							"keys": []interface{}{},
						},
					},
				},
			}},
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
// Key Vault: access policy max items
// ---------------------------------------------------------------------------

func TestKeyVaultValidateAccessPolicyMaxItems(t *testing.T) {
	k := NewKeyVault()

	// Build an array of 1025 policies (over the 1024 limit).
	policies := make([]interface{}, 1025)
	for i := range policies {
		policies[i] = map[string]interface{}{
			"tenantId": "550e8400-e29b-41d4-a716-446655440000",
			"objectId": "550e8400-e29b-41d4-a716-446655440000",
		}
	}

	tests := []struct {
		name      string
		body      map[string]interface{}
		wantCount int
	}{
		{
			"at limit 1024",
			map[string]interface{}{"properties": map[string]interface{}{
				"accessPolicies": policies[:1024],
			}},
			0,
		},
		{
			"over limit 1025",
			map[string]interface{}{"properties": map[string]interface{}{
				"accessPolicies": policies,
			}},
			1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Use ValidateProperties directly; the UUID rules would also fire on 1025 valid UUIDs,
			// but we only count array rule errors here.
			var arrayErrors int
			for _, err := range k.ValidateProperties(tc.body) {
				if contains(err.Error(), "maximum") {
					arrayErrors++
				}
			}
			if arrayErrors != tc.wantCount {
				t.Errorf("got %d array-max errors, want %d", arrayErrors, tc.wantCount)
			}
		})
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchSubstring(s, sub)
}

func searchSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Key Vault: timeouts
// ---------------------------------------------------------------------------

func TestKeyVaultTimeouts(t *testing.T) {
	k := NewKeyVault()
	fallback := 10 * time.Minute

	tests := []struct {
		op   string
		want time.Duration
	}{
		{"create", 30 * time.Minute},
		{"read", 5 * time.Minute},
		{"update", 30 * time.Minute},
		{"delete", 30 * time.Minute},
	}
	for _, tc := range tests {
		t.Run(tc.op, func(t *testing.T) {
			got := k.TimeoutDefault(tc.op, fallback)
			if got != tc.want {
				t.Errorf("TimeoutDefault(%s) = %v, want %v", tc.op, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Key Vault: soft delete
// ---------------------------------------------------------------------------

func TestKeyVaultSoftDelete(t *testing.T) {
	k := NewKeyVault()
	if !k.IsSoftDelete() {
		t.Error("expected IsSoftDelete() = true for Key Vault")
	}
}

// ---------------------------------------------------------------------------
// Key Vault: Validate (combined)
// ---------------------------------------------------------------------------

func TestKeyVaultValidate(t *testing.T) {
	k := NewKeyVault()

	// Valid complete configuration with access policies, network acls, and UUIDs.
	diags := k.Validate("myvault", map[string]interface{}{
		"properties": map[string]interface{}{
			"tenantId": "550e8400-e29b-41d4-a716-446655440000",
			"sku":      map[string]interface{}{"name": "standard", "family": "A"},
			"accessPolicies": []interface{}{
				map[string]interface{}{
					"tenantId": "550e8400-e29b-41d4-a716-446655440000",
					"objectId": "550e8400-e29b-41d4-a716-446655440001",
					"permissions": map[string]interface{}{
						"keys":         []interface{}{"Get", "List"},
						"secrets":      []interface{}{"Get"},
						"certificates": []interface{}{"Get", "List"},
						"storage":      []interface{}{"Get"},
					},
				},
			},
			"networkAcls": map[string]interface{}{
				"defaultAction": "Deny",
				"bypass":        "AzureServices",
				"ipRules": []interface{}{
					map[string]interface{}{"value": "10.0.0.1"},
					map[string]interface{}{"value": "192.168.0.0/16"},
				},
			},
			"softDeleteRetentionInDays": float64(30),
			"publicNetworkAccess":       "Enabled",
			"createMode":                "default",
		},
	}, true)
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics for valid config, got %d:", len(diags))
		for _, d := range diags {
			t.Errorf("  %s: %s", d.Summary(), d.Detail())
		}
	}

	// Invalid name + invalid properties (sku, ip, permission) = 2 diagnostics:
	// 1 "Invalid resource name" + 1 "Invalid property value(s)" (3 errors grouped).
	diags = k.Validate("ab", map[string]interface{}{
		"properties": map[string]interface{}{
			"tenantId": "00000000-0000-0000-0000-000000000000",
			"sku":      map[string]interface{}{"name": "basic", "family": "A"},
			"networkAcls": map[string]interface{}{
				"ipRules": []interface{}{
					map[string]interface{}{"value": "bad-ip"},
				},
			},
			"accessPolicies": []interface{}{
				map[string]interface{}{
					"permissions": map[string]interface{}{
						"keys": []interface{}{"BadPerm"},
					},
				},
			},
		},
	}, true)
	if len(diags) != 2 {
		t.Errorf("expected 2 diagnostics, got %d:", len(diags))
		for _, d := range diags {
			t.Errorf("  %s: %s", d.Summary(), d.Detail())
		}
	}
}

func TestKeyVaultComputedFields(t *testing.T) {
	k := NewKeyVault()
	computed := k.GetComputedFields()
	if len(computed) != 1 || computed[0] != "properties.vaultUri" {
		t.Errorf("unexpected computed fields: %v", computed)
	}
}

func TestKeyVaultDefaultFields(t *testing.T) {
	k := NewKeyVault()

	// GetDefaultFields returns paths derived from DefaultValues.
	defaults := k.GetDefaultFields()
	if len(defaults) != 7 {
		t.Fatalf("expected 7 default fields, got %d: %v", len(defaults), defaults)
	}

	// Verify actual default values are present.
	vals := k.GetDefaultValues()
	byPath := map[string]interface{}{}
	for _, v := range vals {
		byPath[v.PropertyPath] = v.Value
	}
	if v, ok := byPath["properties.enableRbacAuthorization"]; !ok || v != false {
		t.Errorf("expected enableRbacAuthorization default false, got %v", v)
	}
	if v, ok := byPath["properties.networkAcls.defaultAction"]; !ok || v != "Allow" {
		t.Errorf("expected networkAcls.defaultAction default Allow, got %v", v)
	}
	if v, ok := byPath["properties.enableSoftDelete"]; !ok || v != true {
		t.Errorf("expected enableSoftDelete default true, got %v", v)
	}
}

func TestKeyVaultComputedFieldsPackageLevel(t *testing.T) {
	// Test the package-level accessor resolves through the registry.
	EnsureRegistered()
	computed := GetComputedFields("Microsoft.KeyVault/vaults", "")
	if len(computed) == 0 {
		t.Error("expected computed fields from package-level GetComputedFields")
	}
	defaults := GetDefaultFields("Microsoft.KeyVault/vaults", "")
	if len(defaults) == 0 {
		t.Error("expected default fields from package-level GetDefaultFields")
	}
	// Unknown resource returns nil.
	if got := GetComputedFields("Microsoft.Fake/noSuchResource", ""); got != nil {
		t.Errorf("expected nil for unknown resource, got %v", got)
	}
}
