package azwise

import (
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
