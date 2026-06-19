// Command generate-poc is a proof-of-concept that generates the storage account
// resource schema from bicep types.json and writes it to stdout or a file.
//
//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Azure/terraform-provider-azapi/internal/azapin/generator"
	// Registers the per-resource schema customizers (init()) and exposes Apply.
	// Customizers are generation-time only and are not part of the provider runtime.
	"github.com/Azure/terraform-provider-azapi/internal/azapin/generator/customizers"
)

func main() {
	// Find the types.json for storage accounts
	typesPath := filepath.Join("internal", "azure", "generated", "storage", "microsoft.storage", "2025-01-01", "types.json")
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

	// Find storage account
	var sa *generator.ResourceDefinition
	for _, d := range defs {
		if d.Name == "Microsoft.Storage/storageAccounts@2025-01-01" {
			sa = d
			break
		}
	}
	if sa == nil {
		fmt.Fprintf(os.Stderr, "Storage account not found\n")
		os.Exit(1)
	}

	source, err := generator.EmitSchema(sa)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error emitting schema: %v\n", err)
		os.Exit(1)
	}

	// Validate: emitted schema covers all bicep body properties and vice versa.
	// Exclude the synthesized envelope attributes (name / parent reference / id).
	mismatches := generator.ValidateEmittedSchema(source, sa.Body, generator.EnvelopeAttrNames(sa)...)
	if len(mismatches) > 0 {
		fmt.Fprintf(os.Stderr, "Schema validation failed:\n%s", generator.FormatMismatches(mismatches))
		// Count errors (extra and type_mismatch are errors; missing is a warning)
		errors := 0
		for _, m := range mismatches {
			if m.Kind != generator.MismatchMissingInSchema {
				errors++
			}
		}
		if errors > 0 {
			os.Exit(1)
		}
	}

	// Write to file (path includes the service folder, e.g. storage/storage_account_gen.go).
	outPath := filepath.Join("internal", "azapin", "generated", generator.FileName(sa))
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output dir: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(outPath, []byte(source), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", outPath, err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s (%d bytes, %d properties validated)\n", outPath, len(source), countExpectedPaths(sa.Body))
}

func countExpectedPaths(body *generator.Type) int {
	paths := make(map[string]*generator.Property)
	generator.CollectExpectedPaths(body, "", paths)
	return len(paths)
}
