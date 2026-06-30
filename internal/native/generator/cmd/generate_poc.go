// Command generate-poc generates native static resource schemas from the bicep
// types.json embedded under internal/azure/generated and writes one Go file per
// resource into internal/native/generated/<service>/.
//
//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Azure/terraform-provider-azapi/internal/azure"
	"github.com/Azure/terraform-provider-azapi/internal/native/armtypes"
	"github.com/Azure/terraform-provider-azapi/internal/native/generator"
	// Registers the per-resource schema customizers (init()) and exposes Apply.
	// Customizers are generation-time only and are not part of the provider runtime.
	"github.com/Azure/terraform-provider-azapi/internal/native/generator/customizers"
)

// targets lists the ARM resource types to generate. For each, the latest stable
// API version and its types.json location are resolved through the azure schema
// loader (internal/azure), which reads the same embedded index.json/types.json the
// provider runtime loads — so adding a resource is one line here (plus its
// customizer, if any) and no per-namespace directory is hardcoded.
var targets = []string{
	armtypes.StorageAccount,
	armtypes.StorageAccountBlobService,
	armtypes.ResourceGroup,
	armtypes.WebServerFarm,
	armtypes.WebSite,
}

func main() {
	// Resolve each target to its latest-stable version + types.json location via
	// the azure schema loader, then load + process each distinct types.json once
	// (sibling resources such as storageAccounts and its blobServices child share
	// one file). Always picking the latest STABLE version means vendoring a newer
	// index/types.json rolls the generated schema forward without editing this
	// command. types.json is read from the same embedded FS the provider runtime
	// loads, so the generator never reads or reconstructs an on-disk path.
	byName := map[string]*generator.ResourceDefinition{}
	tags := make([]string, 0, len(targets))
	loaded := map[string]bool{}
	for _, armType := range targets {
		version, err := azure.GetLatestStableApiVersion(armType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving %s: %v\n", armType, err)
			os.Exit(1)
		}
		tags = append(tags, armType+"@"+version)

		location, err := azure.GetResourceTypeLocation(armType, version)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving %s: %v\n", armType, err)
			os.Exit(1)
		}
		if loaded[location] {
			continue
		}
		loaded[location] = true

		data, err := azure.StaticFiles.ReadFile("generated/" + location)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", location, err)
			os.Exit(1)
		}
		defs, err := generator.ParseTypesJSON(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", location, err)
			os.Exit(1)
		}
		// Post-process (defaults, validators, azwise overlay, envelope spec), then
		// per-resource developer customizers so hand-written Go has the final say.
		generator.PostProcess(defs)
		customizers.Apply(defs)
		for _, d := range defs {
			byName[d.Name] = d
		}
	}

	for _, tag := range tags {
		def := byName[tag]
		if def == nil {
			fmt.Fprintf(os.Stderr, "resource %s not found in its types.json\n", tag)
			os.Exit(1)
		}
		if err := generate(def); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	}
}

// generate emits, validates, and writes one resource's schema file.
func generate(def *generator.ResourceDefinition) error {
	// Restraint check: framework attribute-flag invariants the bicep types and the
	// framework's startup validation don't catch (Default ⟹ Optional+Computed and
	// not Required/read-only; Default ∈ its own validators).
	if vs := generator.CheckFlagInvariants(def); len(vs) > 0 {
		return fmt.Errorf("schema invariant violations for %s:\n%s", def.Name, generator.FormatViolations(vs))
	}

	source, err := generator.EmitSchema(def)
	if err != nil {
		return fmt.Errorf("emitting %s: %w", def.Name, err)
	}

	// Validate: emitted schema covers all bicep body properties and vice versa.
	// Exclude the synthesized envelope attributes (name / parent reference / id).
	mismatches := generator.ValidateEmittedSchema(source, def.Body, generator.EnvelopeAttrNames(def)...)
	errors := 0
	for _, m := range mismatches {
		// extra and type_mismatch are errors; missing is a warning.
		if m.Kind != generator.MismatchMissingInSchema {
			errors++
		}
	}
	if len(mismatches) > 0 {
		fmt.Fprintf(os.Stderr, "Schema validation for %s:\n%s", def.Name, generator.FormatMismatches(mismatches))
	}
	if errors > 0 {
		return fmt.Errorf("schema validation failed for %s", def.Name)
	}

	// Write to file (path includes the service folder, e.g. storage/storage_account_gen.go).
	outPath := filepath.Join("internal", "native", "generated", generator.FileName(def))
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("creating output dir for %s: %w", def.Name, err)
	}
	if err := os.WriteFile(outPath, []byte(source), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", outPath, err)
	}

	fmt.Printf("Generated %s (%d bytes, %d properties validated)\n", outPath, len(source), countExpectedPaths(def.Body))
	return nil
}

func countExpectedPaths(body *generator.Type) int {
	paths := make(map[string]*generator.Property)
	generator.CollectExpectedPaths(body, "", paths)
	return len(paths)
}
