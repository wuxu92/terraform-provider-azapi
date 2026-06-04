package azwise

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// resetRegistry clears the registry and sync.Once for test isolation.
func resetRegistry(t *testing.T) {
	t.Helper()
	registry = map[string][]ResourceKnowledge{}
	once = sync.Once{}
}

// ---------------------------------------------------------------------------
// Interface compliance
// ---------------------------------------------------------------------------

var _ ResourceKnowledge = (*BaseKnowledge)(nil)

func TestBaseKnowledgeImplementsInterface(t *testing.T) {
	b := &BaseKnowledge{ResourceType: "Microsoft.Test/resources"}
	if b.GetResourceType() != "Microsoft.Test/resources" {
		t.Fatalf("unexpected resource type: %s", b.GetResourceType())
	}
}

// ---------------------------------------------------------------------------
// Registry: Register + Get
// ---------------------------------------------------------------------------

func TestRegistryGetCatchAll(t *testing.T) {
	resetRegistry(t)

	Register(&BaseKnowledge{ResourceType: "Microsoft.Test/things"})

	got := Get("Microsoft.Test/things", "2024-01-01")
	if got == nil {
		t.Fatal("expected catch-all to match, got nil")
	}
	if got.GetResourceType() != "Microsoft.Test/things" {
		t.Fatalf("unexpected type: %s", got.GetResourceType())
	}
}

func TestRegistryGetCaseInsensitive(t *testing.T) {
	resetRegistry(t)

	Register(&BaseKnowledge{ResourceType: "Microsoft.Storage/storageAccounts"})

	got := Get("microsoft.storage/storageaccounts", "any-version")
	if got == nil {
		t.Fatal("expected case-insensitive match, got nil")
	}
}

func TestRegistryGetVersionSpecific(t *testing.T) {
	resetRegistry(t)

	catchAll := &BaseKnowledge{
		ResourceType: "Microsoft.Compute/virtualMachines",
		ForceNew:     []ForceNewRule{{PropertyPath: "properties.hardwareProfile.vmSize"}},
	}
	versionSpecific := &BaseKnowledge{
		ResourceType: "Microsoft.Compute/virtualMachines",
		ApiVersions:  []string{"2024-03-01", "2024-07-01"},
	}

	Register(catchAll)
	Register(versionSpecific)

	// Version-specific should win
	got := Get("Microsoft.Compute/virtualMachines", "2024-03-01")
	if got != versionSpecific {
		t.Fatal("expected version-specific entry")
	}

	// Unknown version should fall through to catch-all
	got = Get("Microsoft.Compute/virtualMachines", "2023-01-01")
	if got != catchAll {
		t.Fatal("expected catch-all entry")
	}

	// Unregistered resource type should return nil
	got = Get("Microsoft.Nonexistent/things", "2024-01-01")
	if got != nil {
		t.Fatal("expected nil for unregistered type")
	}
}

func TestRegistryGetVersionOnlyFallback(t *testing.T) {
	resetRegistry(t)

	versionOnly := &BaseKnowledge{
		ResourceType: "Microsoft.KeyVault/vaults",
		ApiVersions:  []string{"2023-02-01"},
	}
	Register(versionOnly)

	// Exact version match should work.
	got := Get("Microsoft.KeyVault/vaults", "2023-02-01")
	if got != versionOnly {
		t.Fatal("expected exact version match")
	}

	// Empty version should fall back to the only entry.
	got = Get("Microsoft.KeyVault/vaults", "")
	if got != versionOnly {
		t.Fatal("expected fallback to first entry when apiVersion is empty")
	}

	// Mismatched version should also fall back (best-effort).
	got = Get("Microsoft.KeyVault/vaults", "2025-01-01")
	if got != versionOnly {
		t.Fatal("expected fallback to first entry when no exact match")
	}
}

// ---------------------------------------------------------------------------
// BaseKnowledge.CheckForceNew
// ---------------------------------------------------------------------------

func TestBaseKnowledgeCheckForceNew(t *testing.T) {
	b := &BaseKnowledge{
		ResourceType: "Microsoft.Test/things",
		ForceNew:     []ForceNewRule{{PropertyPath: "name"}, {PropertyPath: "properties.sku.tier"}},
	}

	tests := []struct {
		name   string
		old    map[string]interface{}
		new    map[string]interface{}
		expect bool
	}{
		{
			name:   "no change",
			old:    map[string]interface{}{"name": "foo", "properties": map[string]interface{}{"sku": map[string]interface{}{"tier": "Standard"}}},
			new:    map[string]interface{}{"name": "foo", "properties": map[string]interface{}{"sku": map[string]interface{}{"tier": "Standard"}}},
			expect: false,
		},
		{
			name:   "name changed",
			old:    map[string]interface{}{"name": "foo"},
			new:    map[string]interface{}{"name": "bar"},
			expect: true,
		},
		{
			name:   "nested property changed",
			old:    map[string]interface{}{"properties": map[string]interface{}{"sku": map[string]interface{}{"tier": "Standard"}}},
			new:    map[string]interface{}{"properties": map[string]interface{}{"sku": map[string]interface{}{"tier": "Premium"}}},
			expect: true,
		},
		{
			name:   "both nil",
			old:    map[string]interface{}{},
			new:    map[string]interface{}{},
			expect: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := b.CheckForceNew(tc.old, tc.new)
			if got != tc.expect {
				t.Errorf("CheckForceNew = %v, want %v", got, tc.expect)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// StringRule.Validate — name (empty PropertyPath)
// ---------------------------------------------------------------------------

func TestStringRuleValidateName(t *testing.T) {
	rule := &StringRule{
		Regex:     `^[a-z0-9]{3,24}$`,
		MinLength: 3,
		MaxLength: 24,
		Message:   "must be lowercase alphanumeric, 3-24 characters",
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "mystorageacct123", false},
		{"too short", "ab", true},
		{"too long", "abcdefghijklmnopqrstuvwxy", true}, // 25 chars
		{"uppercase", "MyStorage", true},
		{"hyphens", "my-storage", true},
		{"min length exact", "abc", false},
		{"max length exact", "abcdefghijklmnopqrstuvwx", false}, // 24 chars
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := rule.Validate(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// StringRule.Validate — AllowedValues (enum)
// ---------------------------------------------------------------------------

func TestStringRuleAllowedValues(t *testing.T) {
	rule := &StringRule{
		PropertyPath:  "sku.name",
		AllowedValues: []string{"Standard_LRS", "Premium_LRS"},
		Message:       "must be a valid SKU",
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"exact match", "Standard_LRS", false},
		{"case insensitive", "standard_lrs", false},
		{"other allowed", "Premium_LRS", false},
		{"invalid", "Standard_ZZZ", true},
		{"empty", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := rule.Validate(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// StringRule — label()
// ---------------------------------------------------------------------------

func TestStringRuleLabel(t *testing.T) {
	nameRule := &StringRule{}
	if nameRule.label() != "name" {
		t.Errorf("expected 'name', got %q", nameRule.label())
	}
	propRule := &StringRule{PropertyPath: "sku.name"}
	if propRule.label() != "sku.name" {
		t.Errorf("expected 'sku.name', got %q", propRule.label())
	}
}

// ---------------------------------------------------------------------------
// BaseKnowledge.ValidateName with StringRules
// ---------------------------------------------------------------------------

func TestBaseKnowledgeValidateNameWithStringRules(t *testing.T) {
	b := &BaseKnowledge{
		ResourceType: "Microsoft.Test/things",
		StringRules: []StringRule{
			{Regex: `^[a-z]+$`, MinLength: 3, Message: "lowercase only, 3+ chars"},
			{PropertyPath: "sku.name", AllowedValues: []string{"foo"}, Message: "ignored for name"},
		},
	}

	if err := b.ValidateName("abc"); err != nil {
		t.Errorf("valid name should pass: %v", err)
	}
	if err := b.ValidateName("AB"); err == nil {
		t.Error("invalid name should fail")
	}
}

// ---------------------------------------------------------------------------
// BaseKnowledge.ValidateProperties with StringRules + FloatRules
// ---------------------------------------------------------------------------

func TestBaseKnowledgeValidatePropertiesWithStringRules(t *testing.T) {
	b := &BaseKnowledge{
		ResourceType: "Microsoft.Test/things",
		StringRules: []StringRule{
			{Message: "name rule — should be skipped"},
			{PropertyPath: "sku.name", AllowedValues: []string{"Standard", "Premium"}, Message: "valid SKU"},
		},
		FloatRules: []FloatRule{
			{PropertyPath: "properties.count", MinValue: ptr(float64(1)), MaxValue: ptr(float64(10)), Message: "count 1-10"},
		},
	}

	// All valid
	body := map[string]interface{}{
		"sku":        map[string]interface{}{"name": "Standard"},
		"properties": map[string]interface{}{"count": float64(5)},
	}
	if errs := b.ValidateProperties(body); len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}

	// Invalid string + valid numeric
	body = map[string]interface{}{
		"sku":        map[string]interface{}{"name": "Basic"},
		"properties": map[string]interface{}{"count": float64(5)},
	}
	if errs := b.ValidateProperties(body); len(errs) != 1 {
		t.Errorf("expected 1 error (string), got %d: %v", len(errs), errs)
	}

	// Valid string + invalid numeric
	body = map[string]interface{}{
		"sku":        map[string]interface{}{"name": "Premium"},
		"properties": map[string]interface{}{"count": float64(99)},
	}
	if errs := b.ValidateProperties(body); len(errs) != 1 {
		t.Errorf("expected 1 error (numeric), got %d: %v", len(errs), errs)
	}

	// Both invalid
	body = map[string]interface{}{
		"sku":        map[string]interface{}{"name": "Basic"},
		"properties": map[string]interface{}{"count": float64(99)},
	}
	if errs := b.ValidateProperties(body); len(errs) != 2 {
		t.Errorf("expected 2 errors, got %d: %v", len(errs), errs)
	}

	// Missing properties — no errors
	body = map[string]interface{}{}
	if errs := b.ValidateProperties(body); len(errs) != 0 {
		t.Errorf("expected no errors for missing props, got %v", errs)
	}

	// Non-string value at string path — skip
	body = map[string]interface{}{
		"sku": map[string]interface{}{"name": 42},
	}
	if errs := b.ValidateProperties(body); len(errs) != 0 {
		t.Errorf("expected no errors for non-string value, got %v", errs)
	}
}

// ---------------------------------------------------------------------------
// FloatRule.Validate
// ---------------------------------------------------------------------------

func TestFloatRuleValidate(t *testing.T) {
	rule := FloatRule{
		PropertyPath: "properties.ratio",
		MinValue:     ptr(0.5),
		MaxValue:     ptr(2.5),
		Message:      "must be between 0.5 and 2.5",
	}

	tests := []struct {
		name    string
		body    map[string]interface{}
		wantErr bool
	}{
		{"valid", map[string]interface{}{"properties": map[string]interface{}{"ratio": float64(1.0)}}, false},
		{"min boundary", map[string]interface{}{"properties": map[string]interface{}{"ratio": float64(0.5)}}, false},
		{"max boundary", map[string]interface{}{"properties": map[string]interface{}{"ratio": float64(2.5)}}, false},
		{"below min", map[string]interface{}{"properties": map[string]interface{}{"ratio": float64(0.1)}}, true},
		{"above max", map[string]interface{}{"properties": map[string]interface{}{"ratio": float64(3.0)}}, true},
		{"absent property", map[string]interface{}{}, false},
		{"non-numeric", map[string]interface{}{"properties": map[string]interface{}{"ratio": "not-a-number"}}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := rule.Validate(tc.body)
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate(%s) error = %v, wantErr %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// IntRule.Validate
// ---------------------------------------------------------------------------

func TestIntRuleValidate(t *testing.T) {
	rule := IntRule{
		PropertyPath: "properties.retentionDays",
		MinValue:     ptr(int64(7)),
		MaxValue:     ptr(int64(90)),
		Message:      "must be between 7 and 90",
	}

	tests := []struct {
		name    string
		body    map[string]interface{}
		wantErr bool
	}{
		{"valid", map[string]interface{}{"properties": map[string]interface{}{"retentionDays": float64(30)}}, false},
		{"min boundary", map[string]interface{}{"properties": map[string]interface{}{"retentionDays": float64(7)}}, false},
		{"max boundary", map[string]interface{}{"properties": map[string]interface{}{"retentionDays": float64(90)}}, false},
		{"below min", map[string]interface{}{"properties": map[string]interface{}{"retentionDays": float64(3)}}, true},
		{"above max", map[string]interface{}{"properties": map[string]interface{}{"retentionDays": float64(100)}}, true},
		{"absent property", map[string]interface{}{}, false},
		{"non-numeric", map[string]interface{}{"properties": map[string]interface{}{"retentionDays": "not-a-number"}}, false},
		{"fractional rejected", map[string]interface{}{"properties": map[string]interface{}{"retentionDays": float64(7.5)}}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := rule.Validate(tc.body)
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate(%s) error = %v, wantErr %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// BaseKnowledge.Validate (combined method)
// ---------------------------------------------------------------------------

func TestBaseKnowledgeValidate(t *testing.T) {
	b := &BaseKnowledge{
		ResourceType: "Microsoft.Test/things",
		StringRules: []StringRule{
			{Regex: `^[a-z]+$`, MinLength: 3, Message: "lowercase only, 3+ chars"},
			{PropertyPath: "sku.name", AllowedValues: []string{"Standard", "Premium"}, Message: "valid SKU"},
		},
		IntRules: []IntRule{
			{PropertyPath: "properties.count", MinValue: ptr(int64(1)), MaxValue: ptr(int64(10)), Message: "count 1-10"},
		},
		SensitiveFields: []string{"properties.secret"},
	}

	tests := []struct {
		name             string
		resourceName     string
		body             map[string]interface{}
		hasSensitiveBody bool
		wantCount        int
	}{
		{
			"all valid with sensitive_body",
			"abc",
			map[string]interface{}{
				"sku":        map[string]interface{}{"name": "Standard"},
				"properties": map[string]interface{}{"count": float64(5)},
			},
			true,
			0,
		},
		{
			"invalid name only",
			"AB",
			nil,
			false,
			1,
		},
		{
			"invalid property only",
			"",
			map[string]interface{}{
				"sku": map[string]interface{}{"name": "Basic"},
			},
			true,
			1,
		},
		{
			"sensitive field error when no sensitive_body",
			"",
			map[string]interface{}{},
			false,
			1,
		},
		{
			"no sensitive error with sensitive_body",
			"",
			map[string]interface{}{},
			true,
			0,
		},
		{
			"nil body skips property and sensitive checks",
			"",
			nil,
			false,
			0,
		},
		{
			"multiple errors",
			"AB",
			map[string]interface{}{
				"sku":        map[string]interface{}{"name": "Basic"},
				"properties": map[string]interface{}{"count": float64(99)},
			},
			false,
			3, // name + properties (grouped) + sensitive
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			diags := b.Validate(tc.resourceName, tc.body, tc.hasSensitiveBody)
			if len(diags) != tc.wantCount {
				t.Errorf("Validate() got %d diagnostics, want %d:", len(diags), tc.wantCount)
				for _, d := range diags {
					t.Errorf("  [%s] %s: %s", d.Severity(), d.Summary(), d.Detail())
				}
			}
		})
	}
}

func TestBaseKnowledgeValidateDiagnosticSeverity(t *testing.T) {
	b := &BaseKnowledge{
		ResourceType: "Microsoft.Test/things",
		StringRules: []StringRule{
			{Regex: `^[a-z]+$`, MinLength: 3, Message: "lowercase only"},
		},
	}

	diags := b.Validate("AB", nil, false)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Severity() != diag.SeverityError {
		t.Errorf("expected SeverityError, got %v", diags[0].Severity())
	}
	if diags[0].Summary() != "Invalid resource name" {
		t.Errorf("unexpected summary: %s", diags[0].Summary())
	}
}

// ---------------------------------------------------------------------------
// BaseKnowledge.TimeoutDefault
// ---------------------------------------------------------------------------

func TestBaseKnowledgeTimeoutDefault(t *testing.T) {
	b := &BaseKnowledge{
		ResourceType: "Microsoft.Test/things",
		TimeoutsConfig: &Timeouts{
			Create: 60 * time.Minute,
			Delete: 45 * time.Minute,
		},
	}
	fallback := 30 * time.Minute

	if got := b.TimeoutDefault("create", fallback); got != 60*time.Minute {
		t.Errorf("create timeout = %v, want 60m", got)
	}
	if got := b.TimeoutDefault("delete", fallback); got != 45*time.Minute {
		t.Errorf("delete timeout = %v, want 45m", got)
	}
	if got := b.TimeoutDefault("read", fallback); got != fallback {
		t.Errorf("read timeout = %v, want fallback %v", got, fallback)
	}
	if got := b.TimeoutDefault("unknown", fallback); got != fallback {
		t.Errorf("unknown timeout = %v, want fallback %v", got, fallback)
	}

	b2 := &BaseKnowledge{ResourceType: "Microsoft.Test/noTimeouts"}
	if got := b2.TimeoutDefault("create", fallback); got != fallback {
		t.Errorf("nil timeouts create = %v, want fallback %v", got, fallback)
	}
}

// ---------------------------------------------------------------------------
// Package-level functions (through registry)
// ---------------------------------------------------------------------------

func TestPackageLevelFunctionsNoKnowledge(t *testing.T) {
	resetRegistry(t)

	if CheckForceNew("Microsoft.Unknown/things", "v1", nil, nil) {
		t.Error("expected false for unknown resource")
	}
	if diags := Validate("Microsoft.Unknown/things", "v1", "anything", nil, false); diags != nil {
		t.Errorf("expected nil for unknown resource, got: %v", diags)
	}
	if got := TimeoutDefault("Microsoft.Unknown/things", "v1", "create", 30*time.Minute); got != 30*time.Minute {
		t.Errorf("expected fallback, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func TestExtractNestedValue(t *testing.T) {
	m := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": "deep",
			},
		},
		"top": "level",
	}

	if v := extractNestedValue(m, "top"); v != "level" {
		t.Errorf("expected 'level', got %v", v)
	}
	if v := extractNestedValue(m, "a.b.c"); v != "deep" {
		t.Errorf("expected 'deep', got %v", v)
	}
	if v := extractNestedValue(m, "nonexistent"); v != nil {
		t.Errorf("expected nil, got %v", v)
	}
	if v := extractNestedValue(m, "a.nonexistent"); v != nil {
		t.Errorf("expected nil, got %v", v)
	}
}

func TestExtractStringValue(t *testing.T) {
	m := map[string]interface{}{"sku": map[string]interface{}{"name": "Standard_LRS"}}
	if v := extractStringValue(m, "sku.name"); v != "Standard_LRS" {
		t.Errorf("expected Standard_LRS, got %q", v)
	}
	if v := extractStringValue(m, "missing"); v != "" {
		t.Errorf("expected empty string, got %q", v)
	}
}

// ---------------------------------------------------------------------------
// removeNestedField
// ---------------------------------------------------------------------------

func TestRemoveNestedField(t *testing.T) {
	m := map[string]interface{}{
		"properties": map[string]interface{}{
			"vaultUri": "https://vault.azure.net/",
			"sku":      map[string]interface{}{"name": "standard"},
			"nested": map[string]interface{}{
				"deep": map[string]interface{}{
					"value": "keep",
					"gone":  "remove",
				},
			},
		},
		"topLevel": "stays",
	}

	// Remove a nested field
	removeNestedField(m, "properties.vaultUri")
	if _, ok := m["properties"].(map[string]interface{})["vaultUri"]; ok {
		t.Error("vaultUri should have been removed")
	}

	// Remove a deeply nested field
	removeNestedField(m, "properties.nested.deep.gone")
	deep := m["properties"].(map[string]interface{})["nested"].(map[string]interface{})["deep"].(map[string]interface{})
	if _, ok := deep["gone"]; ok {
		t.Error("gone should have been removed")
	}
	if deep["value"] != "keep" {
		t.Error("value should still be present")
	}

	// Removing a non-existent path is a no-op
	removeNestedField(m, "properties.nonexistent.path")
	removeNestedField(m, "totally.missing")

	// Top-level and sku should be intact
	if m["topLevel"] != "stays" {
		t.Error("topLevel should still be present")
	}
	if m["properties"].(map[string]interface{})["sku"].(map[string]interface{})["name"] != "standard" {
		t.Error("sku.name should still be present")
	}
}

// ---------------------------------------------------------------------------
// StripComputedFields
// ---------------------------------------------------------------------------

func TestStripComputedFields(t *testing.T) {
	resetRegistry(t)
	Register(&BaseKnowledge{
		ResourceType: "Microsoft.Test/computed",
		ComputedFields: []string{
			"properties.readOnly",
			"properties.nested.serverGenerated",
		},
	})

	t.Run("strips present fields", func(t *testing.T) {
		body := map[string]interface{}{
			"properties": map[string]interface{}{
				"readOnly": "will-be-removed",
				"userSet":  "stays",
				"nested": map[string]interface{}{
					"serverGenerated": "removed",
					"other":           "stays",
				},
			},
		}
		stripped := StripComputedFields("Microsoft.Test/computed", "", body)
		if !stripped {
			t.Error("expected stripped=true")
		}
		props := body["properties"].(map[string]interface{})
		if _, ok := props["readOnly"]; ok {
			t.Error("readOnly should have been stripped")
		}
		if props["userSet"] != "stays" {
			t.Error("userSet should remain")
		}
		nested := props["nested"].(map[string]interface{})
		if _, ok := nested["serverGenerated"]; ok {
			t.Error("serverGenerated should have been stripped")
		}
		if nested["other"] != "stays" {
			t.Error("other should remain")
		}
	})

	t.Run("no-op when fields absent", func(t *testing.T) {
		body := map[string]interface{}{
			"properties": map[string]interface{}{"userSet": "value"},
		}
		stripped := StripComputedFields("Microsoft.Test/computed", "", body)
		if stripped {
			t.Error("expected stripped=false when no computed fields present")
		}
	})

	t.Run("no-op for unknown resource", func(t *testing.T) {
		body := map[string]interface{}{"anything": "value"}
		stripped := StripComputedFields("Microsoft.Unknown/type", "", body)
		if stripped {
			t.Error("expected stripped=false for unknown resource type")
		}
	})
}

// ---------------------------------------------------------------------------
// Validate — computed field warning
// ---------------------------------------------------------------------------

func TestValidateComputedFieldWarning(t *testing.T) {
	b := &BaseKnowledge{
		ResourceType: "Microsoft.Test/things",
		ComputedFields: []string{
			"properties.vaultUri",
			"properties.primaryEndpoints",
		},
	}

	t.Run("warns when computed fields present in body", func(t *testing.T) {
		body := map[string]interface{}{
			"properties": map[string]interface{}{
				"vaultUri": "https://vault.azure.net/",
				"sku":      map[string]interface{}{"name": "standard"},
			},
		}
		diags := b.Validate("", body, true)
		var warnings int
		for _, d := range diags {
			if d.Severity() == diag.SeverityWarning {
				warnings++
			}
		}
		if warnings != 1 {
			t.Errorf("expected 1 warning, got %d", warnings)
		}
	})

	t.Run("no warning when no computed fields in body", func(t *testing.T) {
		body := map[string]interface{}{
			"properties": map[string]interface{}{
				"sku": map[string]interface{}{"name": "standard"},
			},
		}
		diags := b.Validate("", body, true)
		for _, d := range diags {
			if d.Severity() == diag.SeverityWarning {
				t.Error("expected no warnings")
			}
		}
	})

	t.Run("no warning when body is nil", func(t *testing.T) {
		diags := b.Validate("", nil, true)
		for _, d := range diags {
			if d.Severity() == diag.SeverityWarning {
				t.Error("expected no warnings for nil body")
			}
		}
	})
}

// ---------------------------------------------------------------------------
// Validate — required fields
// ---------------------------------------------------------------------------

func TestValidateRequiredFields(t *testing.T) {
	b := &BaseKnowledge{
		ResourceType: "Microsoft.Test/things",
		RequiredFields: []string{
			"properties.tenantId",
			"properties.sku.name",
			"properties.secret",
		},
		SensitiveFields: []string{
			"properties.secret",
		},
		DefaultValues: []DefaultValue{
			{PropertyPath: "properties.sku.name", Value: "Standard"},
		},
	}

	t.Run("all present", func(t *testing.T) {
		body := map[string]interface{}{
			"properties": map[string]interface{}{
				"tenantId": "tid",
				"sku":      map[string]interface{}{"name": "Premium"},
				"secret":   "s3cret",
			},
		}
		diags := b.Validate("", body, true)
		for _, d := range diags {
			if d.Summary() == "Missing required properties" {
				t.Errorf("should not report missing: %s", d.Detail())
			}
		}
	})

	t.Run("missing non-sensitive field", func(t *testing.T) {
		body := map[string]interface{}{
			"properties": map[string]interface{}{
				"secret": "s3cret",
			},
		}
		diags := b.Validate("", body, true)
		var found bool
		for _, d := range diags {
			if d.Summary() == "Missing required properties" {
				found = true
				if !strings.Contains(d.Detail(), "properties.tenantId") {
					t.Errorf("should mention tenantId: %s", d.Detail())
				}
				if !strings.Contains(d.Detail(), "properties.sku.name") {
					t.Errorf("should mention sku.name: %s", d.Detail())
				}
				// sku.name has a default — should include hint
				if !strings.Contains(d.Detail(), "AzureRM default: Standard") {
					t.Errorf("should include default hint: %s", d.Detail())
				}
			}
		}
		if !found {
			t.Error("expected Missing required properties diagnostic")
		}
	})

	t.Run("sensitive field skipped when sensitive_body set", func(t *testing.T) {
		body := map[string]interface{}{
			"properties": map[string]interface{}{
				"tenantId": "tid",
				"sku":      map[string]interface{}{"name": "Premium"},
			},
		}
		// secret is missing from body but hasSensitiveBody=true → no error for it
		diags := b.Validate("", body, true)
		for _, d := range diags {
			if d.Summary() == "Missing required properties" {
				if strings.Contains(d.Detail(), "properties.secret") {
					t.Errorf("should NOT report sensitive field when sensitive_body is set: %s", d.Detail())
				}
			}
		}
	})

	t.Run("sensitive field reported when no sensitive_body", func(t *testing.T) {
		body := map[string]interface{}{
			"properties": map[string]interface{}{
				"tenantId": "tid",
				"sku":      map[string]interface{}{"name": "Premium"},
			},
		}
		diags := b.Validate("", body, false)
		var found bool
		for _, d := range diags {
			if d.Summary() == "Missing required properties" && strings.Contains(d.Detail(), "properties.secret") {
				found = true
			}
		}
		if !found {
			t.Error("expected sensitive required field to be reported when no sensitive_body")
		}
	})

	t.Run("nil body skips check", func(t *testing.T) {
		diags := b.Validate("", nil, false)
		for _, d := range diags {
			if d.Summary() == "Missing required properties" {
				t.Error("should not check required fields when body is nil")
			}
		}
	})
}
