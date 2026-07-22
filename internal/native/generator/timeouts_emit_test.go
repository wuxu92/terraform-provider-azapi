package generator

import (
	"strings"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/native/typegraph"
)

// TestEmitTimeoutsBakedIntoDescriptor is a white-box guard on the generation
// contract: EmitSchema bakes the per-operation timeouts carried on
// def.Timeouts (populated by ApplyAzwise inside PostProcess) into the emitted
// services.Descriptor as a Timeouts block, adds the "time" import, and renders
// each duration as `N * time.Minute` via renderDuration. It fails if the
// Timeouts block, the "time" import, or any per-op duration is dropped.
func TestEmitTimeoutsBakedIntoDescriptor(t *testing.T) {
	// --- Storage account: Create 60m / Read 5m / Update 60m / Delete 60m ---
	// (source of truth: github.com/wuxu92/azwise/storage_account.go TimeoutsConfig)
	saDefs, saVer := latestStorageDefs(t)
	typegraph.PostProcess(saDefs)
	saSource := emitByTag(t, saDefs, "Microsoft.Storage/storageAccounts@"+saVer)

	// The "time" import must be present because a Timeouts block was emitted.
	if !strings.Contains(saSource, `"time"`) {
		t.Error("missing \"time\" import for the emitted Timeouts block")
	}
	// The Timeouts block itself must be baked into the descriptor.
	if !strings.Contains(saSource, "Timeouts: services.Timeouts{") {
		t.Error("missing Timeouts: services.Timeouts{ block in emitted descriptor")
	}
	// Each per-op duration must render as `N * time.Minute`. gofmt aligns the
	// value column, padding shorter field names (e.g. `Read:`), so compare
	// against a space-normalized copy of the source.
	saNorm := collapseSpaces(saSource)
	for _, want := range []string{
		"Create: 60 * time.Minute,",
		"Read: 5 * time.Minute,",
		"Update: 60 * time.Minute,",
		"Delete: 60 * time.Minute,",
	} {
		if !strings.Contains(saNorm, want) {
			t.Errorf("storage account emitted source missing per-op timeout %q", want)
		}
	}

	// --- Role definition: Create 30m / Read 5m / Update 60m / Delete 30m ---
	// (source of truth: github.com/wuxu92/azwise/role_definition.go TimeoutsConfig)
	// Asserting BOTH create (30m) and update (60m) proves per-op values are
	// preserved distinctly, not collapsed onto a single shared duration.
	rdDefs, rdVer := latestStableDefs(t, "Microsoft.Authorization/roleDefinitions")
	typegraph.PostProcess(rdDefs)
	rdNorm := collapseSpaces(emitByTag(t, rdDefs, "Microsoft.Authorization/roleDefinitions@"+rdVer))

	if !strings.Contains(rdNorm, "Timeouts: services.Timeouts{") {
		t.Error("missing Timeouts: services.Timeouts{ block in role definition descriptor")
	}
	if !strings.Contains(rdNorm, "Create: 30 * time.Minute,") {
		t.Error("role definition emitted source missing Create: 30 * time.Minute")
	}
	if !strings.Contains(rdNorm, "Update: 60 * time.Minute,") {
		t.Error("role definition emitted source missing Update: 60 * time.Minute (per-op update timeout, distinct from create)")
	}
}

// emitByTag selects the def whose Name matches tag and returns its emitted
// schema source, failing the test if the def is missing or emission errors.
func emitByTag(t *testing.T, defs []*typegraph.ResourceDefinition, tag string) string {
	t.Helper()
	var def *typegraph.ResourceDefinition
	for _, d := range defs {
		if d.Name == tag {
			def = d
			break
		}
	}
	if def == nil {
		t.Fatalf("def not found for tag %q", tag)
	}
	source, err := EmitSchema(def)
	if err != nil {
		t.Fatalf("EmitSchema(%s): %v", tag, err)
	}
	return source
}

// collapseSpaces normalizes every run of whitespace (spaces, tabs, newlines) to
// a single space so assertions on struct-field content are robust to gofmt's
// value-column alignment, which pads shorter field names like `Read:`. Each
// `Op: N * time.Minute,` phrase survives intact on one normalized line.
func collapseSpaces(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(s, "\t", " ")), " ")
}
