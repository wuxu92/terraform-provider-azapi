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

	// Apply post-processing: extract defaults from descriptions, promote single-optional children
	generator.PostProcess(defs)

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

	// Write to file
	outDir := filepath.Join("internal", "azapin", "generated")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output dir: %v\n", err)
		os.Exit(1)
	}
	outPath := filepath.Join(outDir, "storage_account.go")
	if err := os.WriteFile(outPath, []byte(source), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", outPath, err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s (%d bytes)\n", outPath, len(source))
}
