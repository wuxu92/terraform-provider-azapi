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

	"github.com/Azure/terraform-provider-azapi/internal/native/armtypes"
	"github.com/Azure/terraform-provider-azapi/internal/native/generator"
	// Registers the per-resource schema customizers (init()) and exposes Apply.
	// Customizers are generation-time only and are not part of the provider runtime.
	"github.com/Azure/terraform-provider-azapi/internal/native/generator/customizers"
)

// storageNamespaceDir holds the embedded bicep types for the Microsoft.Storage
// namespace: one types.json per API version directory (YYYY-MM-DD[-preview]).
const storageNamespaceDir = "internal/azure/generated/storage/microsoft.storage"

func main() {
	// Always generate from the latest STABLE (non-preview) API version available in
	// the embedded types, so vendoring a newer types.json automatically rolls the
	// generated schema forward without editing this command.
	apiVersion, err := generator.LatestStableVersion(storageNamespaceDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving latest stable API version: %v\n", err)
		os.Exit(1)
	}

	// targets lists the resources to generate from this version's types.json, as
	// fully-qualified "<ARMType>@<version>" tags. Add an ARM type here (and its
	// customizer, if any) to generate another resource from the same namespace.
	targets := []string{
		armtypes.StorageAccount + "@" + apiVersion,
		armtypes.StorageAccountBlobService + "@" + apiVersion,
	}

	typesPath := filepath.Join(storageNamespaceDir, apiVersion, "types.json")
	data, err := os.ReadFile(typesPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", typesPath, err)
		os.Exit(1)
	}

	defs, err := generator.ParseTypesJSON(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing types.json: %v\n", err)
		os.Exit(1)
	}

	// Apply post-processing (defaults, validators, azwise overlay, envelope spec),
	// then run per-resource developer customizers so hand-written Go has the final say.
	generator.PostProcess(defs)
	customizers.Apply(defs)

	byName := make(map[string]*generator.ResourceDefinition, len(defs))
	for _, d := range defs {
		byName[d.Name] = d
	}

	for _, tag := range targets {
		def := byName[tag]
		if def == nil {
			fmt.Fprintf(os.Stderr, "resource %s not found in %s\n", tag, typesPath)
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
