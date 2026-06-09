// Package azwise provides per-resource-type operational knowledge for Azure resources.
// It surfaces AzureRM-derived rules (ForceNew, timeouts, naming, property constraints,
// sensitive fields, soft-delete) during terraform plan and validate.
package azwise

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)


// ---------------------------------------------------------------------------
// Interface
// ---------------------------------------------------------------------------

// ResourceKnowledge is the interface each resource type implements.
// Simple resources embed BaseKnowledge in a named struct; complex ones override specific methods.
type ResourceKnowledge interface {
	// GetResourceType returns the Azure resource type (e.g. "Microsoft.Storage/storageAccounts").
	GetResourceType() string

	// GetApiVersions returns API versions this knowledge applies to.
	// Empty means all versions (catch-all).
	GetApiVersions() []string

	// Validate checks the resource name, body properties, and sensitive field usage.
	// name is empty when not configured or unknown; body is nil when not configured or unknown;
	// hasSensitiveBody indicates whether the sensitive_body attribute is set.
	Validate(name string, body map[string]interface{}, hasSensitiveBody bool) diag.Diagnostics

	// CheckForceNew returns true if the body change requires resource replacement.
	CheckForceNew(oldBody, newBody map[string]interface{}) bool

	// TimeoutDefault returns the recommended timeout for the given operation,
	// or fallback if no override is set. operation: "create"|"read"|"update"|"delete".
	TimeoutDefault(operation string, fallback time.Duration) time.Duration

	// IsSoftDelete returns whether the resource supports soft-delete.
	IsSoftDelete() bool

	// GetComputedFields returns ARM property paths that are read-only (server-computed,
	// not present in the ARM Create/Update model). Only fields explicitly marked
	// Computed-only in AzureRM — NOT fields that are merely unsupported by AzureRM.
	GetComputedFields() []string

	// GetDefaultFields returns ARM property paths that have server-side defaults.
	GetDefaultFields() []string

	// GetDefaultValues returns default values from AzureRM for properties with
	// server-side or provider-defined defaults. Value is the ARM-format default.
	GetDefaultValues() []DefaultValue

	// GetRequiredFields returns ARM body property paths that must be present
	// for a valid resource creation request.
	GetRequiredFields() []string
}

// ---------------------------------------------------------------------------
// Supporting types
// ---------------------------------------------------------------------------

// ForceNewRule describes a body property that triggers resource replacement.
type ForceNewRule struct {
	PropertyPath string // dot-separated ARM JSON path (e.g. "sku.name")
}

// DefaultValue pairs an ARM property path with its AzureRM-recommended default.
// Value holds the default in ARM wire format: string for enums, bool for flags,
// float64 for numbers, or nil when AzureRM marks the field Optional+Computed but
// does not define an explicit default.
type DefaultValue struct {
	PropertyPath string      // dot-separated ARM JSON path
	Value        interface{} // ARM-format default (nil = no explicit default, server decides)
}

// StringRule validates a string value — resource name (PropertyPath == "") or body property.
// Supports regex pattern matching, length constraints, and enum (AllowedValues).
type StringRule struct {
	PropertyPath  string   // dot-separated ARM JSON path; empty = resource name attribute
	Regex         string   // pattern the value must match; empty = no pattern check
	MinLength     int      // 0 = no minimum
	MaxLength     int      // 0 = no maximum
	AllowedValues []string // if non-empty, value must be one of these (case-insensitive)
	Message       string   // human-readable explanation
}

// Validate checks a string value against the rule.
func (s *StringRule) Validate(value string) error {
	if s.MinLength > 0 && len(value) < s.MinLength {
		return fmt.Errorf("%s %q is too short (minimum %d characters): %s", s.label(), value, s.MinLength, s.Message)
	}
	if s.MaxLength > 0 && len(value) > s.MaxLength {
		return fmt.Errorf("%s %q is too long (maximum %d characters): %s", s.label(), value, s.MaxLength, s.Message)
	}
	if len(s.AllowedValues) > 0 {
		found := false
		for _, av := range s.AllowedValues {
			if strings.EqualFold(av, value) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%s %q is not a valid value: %s", s.label(), value, s.Message)
		}
	}
	if s.Regex != "" {
		re, err := regexp.Compile(s.Regex)
		if err != nil {
			return fmt.Errorf("invalid regex %q for %s: %s", s.Regex, s.label(), err)
		}
		if !re.MatchString(value) {
			return fmt.Errorf("%s %q does not match pattern: %s", s.label(), value, s.Message)
		}
	}
	return nil
}

func (s *StringRule) label() string {
	if s.PropertyPath == "" {
		return "name"
	}
	return s.PropertyPath
}

// FloatRule validates a numeric body property against float64 min/max bounds.
type FloatRule struct {
	PropertyPath string   // dot-separated ARM JSON path
	MinValue     *float64 // nil = no minimum
	MaxValue     *float64 // nil = no maximum
	Message      string
}

// Validate checks the property value in body against the rule bounds.
func (r *FloatRule) Validate(body map[string]interface{}) error {
	v := extractNestedValue(body, r.PropertyPath)
	if v == nil {
		return nil // property not present; nothing to validate
	}
	f, ok := toFloat64(v)
	if !ok {
		return nil // not numeric; skip
	}
	if r.MinValue != nil && f < *r.MinValue {
		return fmt.Errorf("%s: value %g is below minimum %g: %s", r.PropertyPath, f, *r.MinValue, r.Message)
	}
	if r.MaxValue != nil && f > *r.MaxValue {
		return fmt.Errorf("%s: value %g is above maximum %g: %s", r.PropertyPath, f, *r.MaxValue, r.Message)
	}
	return nil
}

// IntRule validates an integer body property against int64 min/max bounds.
type IntRule struct {
	PropertyPath string // dot-separated ARM JSON path
	MinValue     *int64 // nil = no minimum
	MaxValue     *int64 // nil = no maximum
	Message      string
}

// Validate checks the property value in body against the rule bounds.
// JSON numbers arrive as float64; the value must be a whole number.
func (r *IntRule) Validate(body map[string]interface{}) error {
	v := extractNestedValue(body, r.PropertyPath)
	if v == nil {
		return nil
	}
	f, ok := toFloat64(v)
	if !ok {
		return nil
	}
	i := int64(f)
	if float64(i) != f {
		return fmt.Errorf("%s: value %v must be a whole number: %s", r.PropertyPath, v, r.Message)
	}
	if r.MinValue != nil && i < *r.MinValue {
		return fmt.Errorf("%s: value %d is below minimum %d: %s", r.PropertyPath, i, *r.MinValue, r.Message)
	}
	if r.MaxValue != nil && i > *r.MaxValue {
		return fmt.Errorf("%s: value %d is above maximum %d: %s", r.PropertyPath, i, *r.MaxValue, r.Message)
	}
	return nil
}

// ArrayRule validates structural constraints on an array body property.
type ArrayRule struct {
	PropertyPath string // dot-separated ARM JSON path to the array (no [*])
	MaxItems     int    // 0 = no maximum
	Message      string
}

// Validate checks the array at PropertyPath against the rule constraints.
func (r *ArrayRule) Validate(body map[string]interface{}) error {
	arr := extractArray(body, r.PropertyPath)
	if arr == nil {
		return nil
	}
	if r.MaxItems > 0 && len(arr) > r.MaxItems {
		return fmt.Errorf("%s: array has %d items, maximum is %d: %s", r.PropertyPath, len(arr), r.MaxItems, r.Message)
	}
	return nil
}

// Timeouts overrides default timeout values. Zero means use provider default.
type Timeouts struct {
	Create time.Duration
	Read   time.Duration
	Update time.Duration
	Delete time.Duration
}

// ---------------------------------------------------------------------------
// BaseKnowledge — default implementation
// ---------------------------------------------------------------------------

// BaseKnowledge provides default implementations of ResourceKnowledge from data fields.
// Simple resources embed BaseKnowledge in a named struct; complex ones override specific methods.
type BaseKnowledge struct {
	ResourceType    string
	ApiVersions     []string
	ForceNew        []ForceNewRule
	TimeoutsConfig  *Timeouts
	SoftDelete      bool
	StringRules     []StringRule
	FloatRules      []FloatRule
	IntRules        []IntRule
	ArrayRules      []ArrayRule
	SensitiveFields []string
	// ComputedFields lists ARM property paths that are truly read-only — present in
	// the GET response model but absent from the Create/Update model. Only include
	// fields explicitly Computed-only in AzureRM; fields that are merely unsupported
	// by AzureRM but settable via the ARM API must NOT be listed here, as
	// StripComputedFields would silently discard user-provided values.
	ComputedFields []string
	// DefaultValues pairs ARM property paths with their AzureRM-recommended defaults.
	// Corresponds to Optional+Computed in AzureRM schema: user can set, but Azure
	// (or AzureRM) provides a default if omitted.
	DefaultValues []DefaultValue
	// RequiredFields lists ARM body property paths that must be present for a
	// valid resource creation. Derived from AzureRM Required:true schema fields,
	// excluding envelope properties (name, location, resourceGroup).
	RequiredFields []string
}

func (b *BaseKnowledge) GetResourceType() string         { return b.ResourceType }
func (b *BaseKnowledge) GetApiVersions() []string         { return b.ApiVersions }
func (b *BaseKnowledge) IsSoftDelete() bool               { return b.SoftDelete }
func (b *BaseKnowledge) GetComputedFields() []string      { return b.ComputedFields }
func (b *BaseKnowledge) GetDefaultValues() []DefaultValue { return b.DefaultValues }
func (b *BaseKnowledge) GetRequiredFields() []string      { return b.RequiredFields }

func (b *BaseKnowledge) GetDefaultFields() []string {
	if len(b.DefaultValues) == 0 {
		return nil
	}
	fields := make([]string, len(b.DefaultValues))
	for i, d := range b.DefaultValues {
		fields[i] = d.PropertyPath
	}
	return fields
}

// CheckForceNew compares old and new bodies for changes in ForceNew property paths.
func (b *BaseKnowledge) CheckForceNew(oldBody, newBody map[string]interface{}) bool {
	for _, rule := range b.ForceNew {
		oldVal := extractNestedValue(oldBody, rule.PropertyPath)
		newVal := extractNestedValue(newBody, rule.PropertyPath)
		if !reflect.DeepEqual(oldVal, newVal) {
			return true
		}
	}
	return false
}

// ValidateName validates the resource name against StringRules with empty PropertyPath.
func (b *BaseKnowledge) ValidateName(name string) error {
	for i := range b.StringRules {
		if b.StringRules[i].PropertyPath != "" {
			continue
		}
		if err := b.StringRules[i].Validate(name); err != nil {
			return err
		}
	}
	return nil
}

// ValidateProperties checks FloatRules, IntRules, ArrayRules, and StringRules against body properties.
// StringRules whose PropertyPath contains "[*]" are evaluated against all matching array elements.
func (b *BaseKnowledge) ValidateProperties(body map[string]interface{}) []error {
	var errs []error
	for i := range b.FloatRules {
		if err := b.FloatRules[i].Validate(body); err != nil {
			errs = append(errs, err)
		}
	}
	for i := range b.IntRules {
		if err := b.IntRules[i].Validate(body); err != nil {
			errs = append(errs, err)
		}
	}
	for i := range b.ArrayRules {
		if err := b.ArrayRules[i].Validate(body); err != nil {
			errs = append(errs, err)
		}
	}
	for i := range b.StringRules {
		if b.StringRules[i].PropertyPath == "" {
			continue // name rules are checked via ValidateName
		}
		if strings.Contains(b.StringRules[i].PropertyPath, "[*]") {
			// Array path: validate each extracted string value.
			for _, v := range extractAllValues(body, b.StringRules[i].PropertyPath) {
				s, ok := v.(string)
				if !ok {
					continue
				}
				if err := b.StringRules[i].Validate(s); err != nil {
					errs = append(errs, err)
				}
			}
		} else {
			// Scalar path.
			v := extractNestedValue(body, b.StringRules[i].PropertyPath)
			if v == nil {
				continue
			}
			s, ok := v.(string)
			if !ok {
				continue
			}
			if err := b.StringRules[i].Validate(s); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errs
}

// Validate performs all validation: name, body properties, and sensitive field usage.
// Errors of the same category are grouped into a single Diagnostic to avoid
// repetitive output in terraform plan/apply.
func (b *BaseKnowledge) Validate(name string, body map[string]interface{}, hasSensitiveBody bool) diag.Diagnostics {
	var diags diag.Diagnostics

	if name != "" {
		if err := b.ValidateName(name); err != nil {
			diags.AddError(
				"Invalid resource name",
				fmt.Sprintf("The name %q is not valid: %s", name, err.Error()),
			)
		}
	}

	if body != nil {
		if propErrs := b.ValidateProperties(body); len(propErrs) > 0 {
			var buf strings.Builder
			for i, err := range propErrs {
				if i > 0 {
					buf.WriteByte('\n')
				}
				buf.WriteString("- ")
				buf.WriteString(err.Error())
			}
			diags.AddError(
				"Invalid property value(s)",
				buf.String(),
			)
		}
	}

	if body != nil && len(b.RequiredFields) > 0 {
		// Build a set of sensitive fields so we can skip required-field checks
		// for fields the user may have placed in sensitive_body.
		sensitiveSet := make(map[string]bool, len(b.SensitiveFields))
		for _, sf := range b.SensitiveFields {
			sensitiveSet[sf] = true
		}
		var missing []string
		for _, field := range b.RequiredFields {
			if extractNestedValue(body, field) != nil {
				continue // present in body
			}
			if sensitiveSet[field] && hasSensitiveBody {
				continue // expected in sensitive_body
			}
			missing = append(missing, field)
		}
		if len(missing) > 0 {
			// Append default-value hints when available.
			defaultMap := make(map[string]interface{}, len(b.DefaultValues))
			for _, dv := range b.DefaultValues {
				if dv.Value != nil {
					defaultMap[dv.PropertyPath] = dv.Value
				}
			}
			var buf strings.Builder
			buf.WriteString("The following properties are required for this resource type:\n")
			for _, m := range missing {
				buf.WriteString("- ")
				buf.WriteString(m)
				if v, ok := defaultMap[m]; ok {
					buf.WriteString(fmt.Sprintf(" (AzureRM default: %v)", v))
				}
				buf.WriteByte('\n')
			}
			diags.AddError("Missing required properties", buf.String())
		}
	}

	if body != nil && !hasSensitiveBody && len(b.SensitiveFields) > 0 {
		diags.AddError(
			"Sensitive properties should use sensitive_body",
			fmt.Sprintf("Known sensitive fields (%s) should be moved from body to sensitive_body.",
				strings.Join(b.SensitiveFields, ", ")),
		)
	}

	if body != nil && len(b.ComputedFields) > 0 {
		var found []string
		for _, field := range b.ComputedFields {
			if extractNestedValue(body, field) != nil {
				found = append(found, field)
			}
		}
		if len(found) > 0 {
			diags.AddWarning(
				"Read-only properties in body will be ignored",
				fmt.Sprintf("The following properties are computed by the server and cannot be set: %s. "+
					"Remove them from body to avoid perpetual diffs.",
					strings.Join(found, ", ")),
			)
		}
	}

	return diags
}

// TimeoutDefault returns the configured timeout for the operation, or fallback.
func (b *BaseKnowledge) TimeoutDefault(operation string, fallback time.Duration) time.Duration {
	if b.TimeoutsConfig == nil {
		return fallback
	}
	switch operation {
	case "create":
		if b.TimeoutsConfig.Create > 0 {
			return b.TimeoutsConfig.Create
		}
	case "read":
		if b.TimeoutsConfig.Read > 0 {
			return b.TimeoutsConfig.Read
		}
	case "update":
		if b.TimeoutsConfig.Update > 0 {
			return b.TimeoutsConfig.Update
		}
	case "delete":
		if b.TimeoutsConfig.Delete > 0 {
			return b.TimeoutsConfig.Delete
		}
	}
	return fallback
}

// ---------------------------------------------------------------------------
// Registry
// ---------------------------------------------------------------------------

var (
	registry = map[string][]ResourceKnowledge{}
	once     sync.Once
)

// Register adds a ResourceKnowledge implementation to the registry.
func Register(k ResourceKnowledge) {
	key := strings.ToLower(k.GetResourceType())
	registry[key] = append(registry[key], k)
}

// Get returns the best-matching ResourceKnowledge for a resource type + API version.
// Priority: exact version match > catch-all (empty ApiVersions) > first registered entry
// (when apiVersion is empty or no exact match exists).
func Get(resourceType, apiVersion string) ResourceKnowledge {
	key := strings.ToLower(resourceType)
	entries := registry[key]
	if len(entries) == 0 {
		return nil
	}
	var catchAll ResourceKnowledge
	for _, k := range entries {
		versions := k.GetApiVersions()
		if len(versions) == 0 {
			catchAll = k
			continue
		}
		for _, v := range versions {
			if v == apiVersion {
				return k
			}
		}
	}
	if catchAll != nil {
		return catchAll
	}
	// No exact match and no catch-all: return the first entry as a best-effort
	// fallback. This handles the case where apiVersion is "" (unknown) or
	// the caller uses a version not listed but the knowledge is still applicable.
	return entries[0]
}

// EnsureRegistered calls RegisterAll exactly once.
func EnsureRegistered() {
	once.Do(RegisterAll)
}

// ---------------------------------------------------------------------------
// Package-level functions (called from azapi_resource.go)
// ---------------------------------------------------------------------------

// Validate performs all validation checks for a resource type.
func Validate(resourceType, apiVersion, name string, body map[string]interface{}, hasSensitiveBody bool) diag.Diagnostics {
	EnsureRegistered()
	k := Get(resourceType, apiVersion)
	if k == nil {
		return nil
	}
	return k.Validate(name, body, hasSensitiveBody)
}

// CheckForceNew returns true if the body change requires resource replacement
// for the given resource type and API version.
func CheckForceNew(resourceType, apiVersion string, oldBody, newBody map[string]interface{}) bool {
	EnsureRegistered()
	k := Get(resourceType, apiVersion)
	if k == nil {
		return false
	}
	return k.CheckForceNew(oldBody, newBody)
}

// TimeoutDefault returns the recommended timeout for the given operation.
func TimeoutDefault(resourceType, apiVersion, operation string, fallback time.Duration) time.Duration {
	EnsureRegistered()
	k := Get(resourceType, apiVersion)
	if k == nil {
		return fallback
	}
	return k.TimeoutDefault(operation, fallback)
}

// GetComputedFields returns ARM property paths that are read-only (server-computed)
// for the given resource type and API version.
func GetComputedFields(resourceType, apiVersion string) []string {
	EnsureRegistered()
	k := Get(resourceType, apiVersion)
	if k == nil {
		return nil
	}
	return k.GetComputedFields()
}

// GetDefaultFields returns ARM property paths that have server-side defaults
// for the given resource type and API version.
func GetDefaultFields(resourceType, apiVersion string) []string {
	EnsureRegistered()
	k := Get(resourceType, apiVersion)
	if k == nil {
		return nil
	}
	return k.GetDefaultFields()
}

// GetDefaultValues returns default values from AzureRM knowledge for properties
// with known defaults, for the given resource type and API version.
func GetDefaultValues(resourceType, apiVersion string) []DefaultValue {
	EnsureRegistered()
	k := Get(resourceType, apiVersion)
	if k == nil {
		return nil
	}
	return k.GetDefaultValues()
}

// GetRequiredFields returns ARM body property paths that must be present
// for a valid resource creation, for the given resource type and API version.
func GetRequiredFields(resourceType, apiVersion string) []string {
	EnsureRegistered()
	k := Get(resourceType, apiVersion)
	if k == nil {
		return nil
	}
	return k.GetRequiredFields()
}

// StripComputedFields removes known read-only (server-computed) properties from body.
// Returns true if any field was present and removed. This prevents perpetual plan
// diffs when users include server-computed properties in their config body.
func StripComputedFields(resourceType, apiVersion string, body map[string]interface{}) bool {
	EnsureRegistered()
	k := Get(resourceType, apiVersion)
	if k == nil {
		return false
	}
	stripped := false
	for _, field := range k.GetComputedFields() {
		if extractNestedValue(body, field) != nil {
			removeNestedField(body, field)
			stripped = true
		}
	}
	return stripped
}
