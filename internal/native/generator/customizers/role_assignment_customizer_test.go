package customizers_test

import (
	"strings"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/azure"
	"github.com/Azure/terraform-provider-azapi/internal/native/generator"
	"github.com/Azure/terraform-provider-azapi/internal/native/typegraph"
)

// roleAssignmentDef loads the real latest-stable roleAssignments graph through the
// generation pipeline (runtime graph + customizers), matching the source consumed by
// the generated azapi_role_assignment schema.
func roleAssignmentDef(t *testing.T) *typegraph.ResourceDefinition {
	t.Helper()
	const armType = "Microsoft.Authorization/roleAssignments"

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

// TestRoleAssignmentSchemaEmitsUUIDRequiredForceNewContract guards the generated
// static-resource schema for role assignments: callers get azapi_role_assignment under
// scope_id, the envelope name is a UUID, ARM-required body fields are Required, and
// every AzureRM immutable role-assignment body field gets RequiresReplace.
func TestRoleAssignmentSchemaEmitsUUIDRequiredForceNewContract(t *testing.T) {
	def := roleAssignmentDef(t)

	src, err := generator.EmitSchema(def)
	if err != nil {
		t.Fatalf("EmitSchema: %v", err)
	}

	descriptor := sourceAttributeBlock(t, src, "var RoleAssignment = services.Descriptor{")
	for _, want := range []string{`"azapi_role_assignment"`, `"Microsoft.Authorization/roleAssignments"`, `"2022-04-01"`, `"scope_id"`} {
		if !strings.Contains(descriptor, want) {
			t.Errorf("RoleAssignment descriptor missing %q", want)
		}
	}

	name := sourceAttributeBlock(t, src, `"name": schema.StringAttribute{`)
	for _, want := range []string{"validators.UUID()", "stringplanmodifier.RequiresReplace()"} {
		if !strings.Contains(name, want) {
			t.Errorf("name block missing %q", want)
		}
	}

	for _, attr := range []string{`"role_definition_id": schema.StringAttribute{`, `"principal_id": schema.StringAttribute{`} {
		block := sourceAttributeBlock(t, src, attr)
		assertRequiredOnlyAttribute(t, attr, block)
		if !strings.Contains(block, "stringplanmodifier.RequiresReplace()") {
			t.Errorf("%s block missing RequiresReplace", attr)
		}
	}

	for _, attr := range []string{
		`"principal_type": schema.StringAttribute{`,
		`"delegated_managed_identity_resource_id": schema.StringAttribute{`,
		`"description": schema.StringAttribute{`,
		`"condition": schema.StringAttribute{`,
		`"condition_version": schema.StringAttribute{`,
	} {
		block := sourceAttributeBlock(t, src, attr)
		if !strings.Contains(block, "stringplanmodifier.RequiresReplace()") {
			t.Errorf("%s block missing RequiresReplace", attr)
		}
	}

	delegatedID := sourceAttributeBlock(t, src, `"delegated_managed_identity_resource_id": schema.StringAttribute{`)
	if !strings.Contains(delegatedID, "validators.AzureResourceID()") {
		t.Error("delegated_managed_identity_resource_id block missing AzureResourceID validator")
	}
}

func roleDefinitionDef(t *testing.T) *typegraph.ResourceDefinition {
	t.Helper()
	const armType = "Microsoft.Authorization/roleDefinitions"

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

func TestRoleDefinitionSchemaEmitsScopeID(t *testing.T) {
	def := roleDefinitionDef(t)
	src, err := generator.EmitSchema(def)
	if err != nil {
		t.Fatalf("EmitSchema: %v", err)
	}

	descriptor := sourceAttributeBlock(t, src, "var RoleDefinition = services.Descriptor{")
	for _, want := range []string{`"azapi_role_definition"`, `"Microsoft.Authorization/roleDefinitions"`, `"2022-04-01"`, `"scope_id"`} {
		if !strings.Contains(descriptor, want) {
			t.Errorf("RoleDefinition descriptor missing %q", want)
		}
	}
	if !strings.Contains(src, `"scope_id": schema.StringAttribute{`) {
		t.Error("role definition schema missing scope_id envelope attribute")
	}
}
