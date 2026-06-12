// Command azapin-validate validates compiled azapin schemas against their
// source bicep type definitions. Each schema self-describes its resource type
// via an [azapin:ResourceType@Version] tag in the Description field.
//
// Usage:
//
//	go run ./internal/azapin/cmd/azapin-validate/                       # validate all
//	go run ./internal/azapin/cmd/azapin-validate/ -r azapi_storage_account  # validate one

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Azure/terraform-provider-azapi/internal/azapin/generator"
	"github.com/Azure/terraform-provider-azapi/internal/azapin/validate"

	// Import generated package to trigger init() registrations
	"github.com/Azure/terraform-provider-azapi/internal/azapin/generated"
)

func main() {
	res := flag.String("r", "", "resource name to validate (e.g. azapi_storage_account); omit to validate all")
	flag.StringVar(res, "res", "", "resource name to validate (e.g. azapi_storage_account); omit to validate all")
	flag.Parse()

	// Determine which schemas to validate
	targets := generated.Registry
	if *res != "" {
		fn, ok := generated.Registry[*res]
		if !ok {
			fmt.Fprintf(os.Stderr, "Unknown resource: %s\nAvailable:\n", *res)
			for k := range generated.Registry {
				fmt.Fprintf(os.Stderr, "  %s\n", k)
			}
			os.Exit(1)
		}
		targets = map[string]generated.SchemaFunc{*res: fn}
	}

	// Resolve the project root from this source file's location:
	// this file is at internal/azapin/cmd/azapin-validate/main.go
	// project root is 4 levels up
	projectRoot := resolveProjectRoot()

	// Load the index file for types.json path resolution
	indexPath := filepath.Join(projectRoot, "internal", "azure", "generated", "index.json")
	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading index.json at %s: %v\n", indexPath, err)
		os.Exit(1)
	}

	var index struct {
		Resources map[string]json.RawMessage `json:"resources"`
	}
	if err := json.Unmarshal(indexData, &index); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing index.json: %v\n", err)
		os.Exit(1)
	}

	totalErrors := 0
	totalWarnings := 0
	validated := 0

	for name, fn := range targets {
		s := fn()

		// Extract resource tag from Description
		tag, ok := validate.ExtractResourceTag(s.Description)
		if !ok {
			fmt.Fprintf(os.Stderr, "SKIP %s: no [azapin:...] tag in Description\n", name)
			continue
		}

		// Find the types.json path from the index
		typesPath, err := resolveTypesPath(tag, index.Resources, projectRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "SKIP %s: %v\n", name, err)
			continue
		}

		// Parse the bicep types
		typesData, err := os.ReadFile(typesPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR %s: reading %s: %v\n", name, typesPath, err)
			totalErrors++
			continue
		}

		defs, err := generator.ParseTypesJSON(typesData)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR %s: parsing types.json: %v\n", name, err)
			totalErrors++
			continue
		}

		// Find the matching resource definition
		var body *generator.Type
		for _, def := range defs {
			if def.Name == tag {
				body = def.Body
				break
			}
		}
		if body == nil {
			fmt.Fprintf(os.Stderr, "ERROR %s: resource %s not found in %s\n", name, tag, typesPath)
			totalErrors++
			continue
		}

		// Apply same post-processing as the generator
		generator.PostProcess([]*generator.ResourceDefinition{{Name: tag, Body: body}})

		// Validate
		mismatches := validate.SchemaAgainstBicep(s, body)

		errors := 0
		warnings := 0
		for _, m := range mismatches {
			switch m.Kind {
			case validate.MismatchExtraInSchema, validate.MismatchTypeMismatch:
				errors++
			case validate.MismatchMissingInSchema:
				warnings++
			}
		}

		if errors > 0 || warnings > 0 {
			fmt.Fprintf(os.Stderr, "FAIL %s (%s): %s", name, tag, validate.FormatMismatches(mismatches))
		} else {
			fmt.Printf("OK   %s (%s): 0 mismatches\n", name, tag)
		}

		totalErrors += errors
		totalWarnings += warnings
		validated++
	}

	fmt.Printf("\n%d schemas validated, %d errors, %d warnings\n", validated, totalErrors, totalWarnings)
	if totalErrors > 0 {
		os.Exit(1)
	}
}

// resolveProjectRoot determines the project root directory from this source
// file's location using runtime.Caller. This file lives at:
//
//	<project>/internal/azapin/cmd/azapin-validate/main.go
//
// So the project root is 4 directories up.
func resolveProjectRoot() string {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: cannot determine source file location via runtime.Caller\n")
		os.Exit(1)
	}
	// thisFile = .../internal/azapin/cmd/azapin-validate/main.go
	// Go up 4 levels: azapin-validate → cmd → azapin → internal → project root
	root := filepath.Dir(thisFile)
	for i := 0; i < 4; i++ {
		root = filepath.Dir(root)
	}
	return root
}

// resolveTypesPath finds the types.json file for a given "ResourceType@Version" tag.
func resolveTypesPath(tag string, resources map[string]json.RawMessage, projectRoot string) (string, error) {
	// Check if the tag exists directly in the index
	if _, ok := resources[tag]; ok {
		return tagToTypesPath(tag, projectRoot)
	}

	// Try case-insensitive match
	tagLower := strings.ToLower(tag)
	for key := range resources {
		if strings.ToLower(key) == tagLower {
			return tagToTypesPath(key, projectRoot)
		}
	}

	return "", fmt.Errorf("resource %q not found in index.json", tag)
}

func tagToTypesPath(tag string, projectRoot string) (string, error) {
	// Split: Microsoft.Storage/storageAccounts@2025-01-01
	parts := strings.SplitN(tag, "@", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid tag format: %s (expected ResourceType@Version)", tag)
	}
	resourceType := parts[0] // Microsoft.Storage/storageAccounts
	apiVersion := parts[1]   // 2025-01-01

	// Split resource type: Microsoft.Storage/storageAccounts → Microsoft.Storage, storageAccounts
	rtParts := strings.SplitN(resourceType, "/", 2)
	if len(rtParts) < 2 {
		return "", fmt.Errorf("invalid resource type: %s", resourceType)
	}
	namespace := strings.ToLower(rtParts[0]) // microsoft.storage

	// Service directory: microsoft.storage → storage
	nsParts := strings.Split(namespace, ".")
	service := nsParts[len(nsParts)-1] // storage

	path := filepath.Join(projectRoot, "internal", "azure", "generated", service, namespace, apiVersion, "types.json")
	return path, nil
}
