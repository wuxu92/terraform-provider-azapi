package customizers_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/azure"
	"github.com/Azure/terraform-provider-azapi/internal/native/generator"
	"github.com/Azure/terraform-provider-azapi/internal/native/generator/customizers"
	"github.com/Azure/terraform-provider-azapi/internal/native/typegraph"
)

// keyVaultDef loads the real latest-stable Key Vault bicep defs from the same
// embedded source the provider ships, runs PostProcess + customizers.Apply (so
// the graph reflects exactly what the emitter consumes), and returns the vault
// definition selected by ARM type. It replicates the loader body from
// generator/fixtures_test.go (unexported there, and generator can't import
// customizers without a cycle). Skips when the version/types.json can't be
// resolved; fatals on a genuine parse failure.
func keyVaultDef(t *testing.T) *typegraph.ResourceDefinition {
	t.Helper()
	const armType = "Microsoft.KeyVault/vaults"

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

	typegraph.PostProcess(defs)
	customizers.Apply(defs)

	for _, def := range defs {
		if typegraph.ARMTypeOf(def) == armType {
			return def
		}
	}
	t.Fatalf("no %s def in parsed types (version %s)", armType, version)
	return nil
}

// keyVaultKeyDef loads the real latest-stable Key Vault key bicep defs and runs
// the same generation graph builder used by generated _gen.go files. The test
// below asserts emitted source because the generated schema is the compatibility
// surface consumed by native resource registration.
func keyVaultKeyDef(t *testing.T) *typegraph.ResourceDefinition {
	t.Helper()
	const armType = "Microsoft.KeyVault/vaults/keys"

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

	defs, err := generator.BuildForGeneration(data)
	if err != nil {
		t.Fatalf("BuildForGeneration: %v", err)
	}
	for _, def := range defs {
		if typegraph.ARMTypeOf(def) == armType {
			return def
		}
	}
	t.Fatalf("no %s def in generated graph (version %s)", armType, version)
	return nil
}

func sourceAttributeBlock(t *testing.T, src, marker string) string {
	t.Helper()
	start := strings.Index(src, marker)
	if start == -1 {
		t.Fatalf("emitted source missing attribute marker %q", marker)
	}
	open := strings.Index(src[start:], "{")
	if open == -1 {
		t.Fatalf("emitted source attribute marker %q has no opening brace", marker)
	}

	depth := 0
	for i := start + open; i < len(src); i++ {
		switch src[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return src[start : i+1]
			}
		}
	}
	t.Fatalf("emitted source attribute marker %q has no closing brace", marker)
	return ""
}

func assertRequiredOnlyAttribute(t *testing.T, name, block string) {
	t.Helper()

	if !attributeBlockContainsTrueField(block, "Required:") {
		t.Errorf("%s block is not required", name)
	}
	for _, forbidden := range []string{"Optional:", "Computed:"} {
		if strings.Contains(block, forbidden) {
			t.Errorf("%s block contains %s", name, forbidden)
		}
	}
}

func attributeBlockContainsTrueField(block, field string) bool {
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, field) && strings.Contains(line, "true,") {
			return true
		}
	}
	return false
}

// TestKeyVaultKeySchemaEmitsManagementPlaneKeyOpsValidator guards the generated
// schema contract for Key Vault keys: callers get the native azapi_key_vault_key
// resource under key_vault_id, create-time key material fields are required, and
// the key_ops list validates the ARM management-plane operations. It catches the
// past drift where data-plane operations leaked into keyOps and management-plane
// import/release were absent.
func TestKeyVaultKeySchemaEmitsManagementPlaneKeyOpsValidator(t *testing.T) {
	def := keyVaultKeyDef(t)

	src, err := generator.EmitSchema(def)
	if err != nil {
		t.Fatalf("EmitSchema: %v", err)
	}

	descriptor := sourceAttributeBlock(t, src, "var KeyVaultKey = services.Descriptor{")
	for _, want := range []string{`"azapi_key_vault_key"`, `"key_vault_id"`} {
		if !strings.Contains(descriptor, want) {
			t.Errorf("KeyVaultKey descriptor missing %q", want)
		}
	}

	kty := sourceAttributeBlock(t, src, `"kty": schema.StringAttribute{`)
	keyOps := sourceAttributeBlock(t, src, `"key_ops": schema.ListAttribute{`)
	for _, attr := range []struct {
		name  string
		block string
	}{
		{name: "kty", block: kty},
		{name: "key_ops", block: keyOps},
	} {
		assertRequiredOnlyAttribute(t, attr.name, attr.block)
	}

	for _, want := range []string{
		"Validators: []validator.List{",
		"listvalidator.ValueStringsAre(",
		"stringvalidator.OneOf(",
		`"import"`,
		`"release"`,
	} {
		if !strings.Contains(keyOps, want) {
			t.Errorf("key_ops block missing %q", want)
		}
	}
	for _, blocked := range []string{`"backup"`, `"delete"`} {
		if strings.Contains(keyOps, blocked) {
			t.Errorf("key_ops block contains data-plane operation %s", blocked)
		}
	}
}

// keyVaultPermissionSets is the ordered, per-list allowed value set the
// customizer attaches to each access-policy permission list. Order matters — the
// customizer preserves it, and the graph/emit assertions below are order-sensitive.
var keyVaultPermissionSets = map[string][]string{
	"properties.accessPolicies.permissions.certificates": {
		"Backup", "Create", "Delete", "DeleteIssuers", "Get", "GetIssuers",
		"Import", "List", "ListIssuers", "ManageContacts", "ManageIssuers",
		"Purge", "Recover", "Restore", "SetIssuers", "Update",
	},
	"properties.accessPolicies.permissions.keys": {
		"Backup", "Create", "Decrypt", "Delete", "Encrypt", "Get", "Import",
		"List", "Purge", "Recover", "Restore", "Sign", "UnwrapKey", "Update",
		"Verify", "WrapKey", "Release", "Rotate", "GetRotationPolicy",
		"SetRotationPolicy",
	},
	"properties.accessPolicies.permissions.secrets": {
		"Backup", "Delete", "Get", "List", "Purge", "Recover", "Restore", "Set",
	},
	"properties.accessPolicies.permissions.storage": {
		"Backup", "Delete", "DeleteSAS", "Get", "GetSAS", "List", "ListSAS",
		"Purge", "Recover", "RegenerateKey", "Restore", "Set", "SetSAS", "Update",
	},
}

// TestKeyVaultCustomizerAttachesPermissionValidators guards the customizer
// contract on the post-Apply graph: each of the four access-policy permission
// lists must carry exactly one case-insensitive one-of validator whose allowed
// set equals the expected ordered slice. Fails loudly if the customizer stops
// attaching a set or an allowed set drifts (order included).
func TestKeyVaultCustomizerAttachesPermissionValidators(t *testing.T) {
	def := keyVaultDef(t)

	for path, want := range keyVaultPermissionSets {
		p := typegraph.FindProperty(def, path)
		if p == nil {
			t.Errorf("%s: property not found", path)
			continue
		}

		var oneOf []typegraph.DescriptionValidator
		for _, v := range p.Validators {
			if v.Kind == typegraph.ValidatorStringOneOfCaseInsensitive {
				oneOf = append(oneOf, v)
			}
		}
		if len(oneOf) != 1 {
			t.Errorf("%s: got %d case-insensitive one-of validators, want exactly 1 (validators=%#v)",
				path, len(oneOf), p.Validators)
			continue
		}
		if !reflect.DeepEqual(oneOf[0].Allowed, want) {
			t.Errorf("%s: allowed set drift\n got: %#v\nwant: %#v", path, oneOf[0].Allowed, want)
		}
	}
}

// TestKeyVaultSchemaEmitsPermissionListValidators guards the emit contract: the
// generated schema source must wrap each permission list's element validator in a
// listvalidator, emit exactly four case-insensitive one-of validators (one per
// list), and keep each list's values within its own list. The distinctive-value
// counts (each unique to one permission list) catch a value leaking across
// siblings — a leak would bump the sibling's count above one.
func TestKeyVaultSchemaEmitsPermissionListValidators(t *testing.T) {
	def := keyVaultDef(t)

	src, err := generator.EmitSchema(def)
	if err != nil {
		t.Fatalf("EmitSchema: %v", err)
	}

	if !strings.Contains(src, "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator") {
		t.Error("emitted source missing listvalidator import path")
	}
	if got := strings.Count(src, "stringvalidator.OneOfCaseInsensitive("); got != 4 {
		t.Errorf("stringvalidator.OneOfCaseInsensitive( count = %d, want 4 (one per permission list)", got)
	}
	if !strings.Contains(src, "listvalidator.ValueStringsAre(") {
		t.Error("emitted source missing listvalidator.ValueStringsAre( wrapper")
	}

	// Each value is unique to exactly one permission list (DeleteIssuers →
	// certificates, WrapKey → keys, GetSAS → storage), so a leak into a sibling
	// list would push its count above one.
	for value, want := range map[string]int{
		"DeleteIssuers": 1,
		"WrapKey":       1,
		"GetSAS":        1,
	} {
		if got := strings.Count(src, value); got != want {
			t.Errorf("value %q appears %d times, want %d (a leak across permission lists?)", value, got, want)
		}
	}
}

// TestKeyVaultCustomizerFlagsEmptyListDefaults guards the customizer contract on
// the post-Apply graph: both networkAcls array properties (ipRules,
// virtualNetworkRules) must carry the DefaultEmptyList flag so the emitter emits
// an empty-list schema Default, while the sibling scalar bypass (a StaticString
// default, not an empty-list default) must not. Fails loudly if the flag stops
// being set on a target list or leaks onto the control sibling.
func TestKeyVaultCustomizerFlagsEmptyListDefaults(t *testing.T) {
	def := keyVaultDef(t)

	for _, path := range []string{
		"properties.networkAcls.ipRules",
		"properties.networkAcls.virtualNetworkRules",
	} {
		p := typegraph.FindProperty(def, path)
		if p == nil {
			t.Errorf("%s: property not found", path)
			continue
		}
		if !p.DefaultEmptyList {
			t.Errorf("%s: DefaultEmptyList = false, want true", path)
		}
	}

	const control = "properties.networkAcls.bypass"
	if p := typegraph.FindProperty(def, control); p == nil {
		t.Errorf("%s: property not found", control)
	} else if p.DefaultEmptyList {
		t.Errorf("%s: DefaultEmptyList = true, want false (control sibling)", control)
	}
}

// TestKeyVaultSchemaEmitsEmptyListDefaults guards the emit contract for the
// empty-list defaults: the generated schema source must import listdefault + attr,
// emit exactly two listdefault.StaticValue(types.ListValueMust( defaults (one per
// flagged list), and carry each list's element-type literal fragments. Catches the
// emitter dropping a default or emitting the wrong element type. The element-type
// fragments are asserted per key/type to stay robust against gofmt spacing.
func TestKeyVaultSchemaEmitsEmptyListDefaults(t *testing.T) {
	def := keyVaultDef(t)

	src, err := generator.EmitSchema(def)
	if err != nil {
		t.Fatalf("EmitSchema: %v", err)
	}

	if !strings.Contains(src, "github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault") {
		t.Error("emitted source missing listdefault import path")
	}
	if !strings.Contains(src, "github.com/hashicorp/terraform-plugin-framework/attr") {
		t.Error("emitted source missing attr import path")
	}
	if got := strings.Count(src, "listdefault.StaticValue(types.ListValueMust("); got != 2 {
		t.Errorf("listdefault.StaticValue(types.ListValueMust( count = %d, want 2 (one per empty-list default)", got)
	}

	// ip_rules element type is {value}; virtual_network_rules is
	// {id, ignore_missing_vnet_service_endpoint} (deterministic sorted key order).
	for _, frag := range []string{
		`"value": types.StringType`,
		`"id": types.StringType`,
		`"ignore_missing_vnet_service_endpoint": types.BoolType`,
	} {
		if !strings.Contains(src, frag) {
			t.Errorf("emitted source missing element-type fragment %q", frag)
		}
	}
}
