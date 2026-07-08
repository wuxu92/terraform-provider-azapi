// Command generate-poc generates native static resource schemas from the bicep
// types.json embedded under internal/azure/generated and writes one Go file per
// resource into internal/native/services/<service>/.
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
	"github.com/Azure/terraform-provider-azapi/internal/native/typegraph"
	// Registers the per-resource schema customizers (init()) and exposes Apply.
	// Customizers are generation-time only and are not part of the provider runtime.
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
	armtypes.UserAssignedIdentity,
	armtypes.KeyVault,
	armtypes.RoleDefinition,
}

func main() {
	// Resolve each target to its latest-stable version + types.json location via
	// the azure schema loader, then load + process each distinct types.json once
	// (sibling resources such as storageAccounts and its blobServices child share
	// one file). Always picking the latest STABLE version means vendoring a newer
	// index/types.json rolls the generated schema forward without editing this
	// command. types.json is read from the same embedded FS the provider runtime
	// loads, so the generator never reads or reconstructs an on-disk path.
	byName := map[string]*typegraph.ResourceDefinition{}
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
		defs, err := generator.BuildForGeneration(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error building %s: %v\n", location, err)
			os.Exit(1)
		}
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

// generate runs the generation pipeline for one resource and writes its schema
// file. Emission and self-verification are owned by generator.Generate; this
// command keeps only target selection, file I/O, and warning reporting.
func generate(def *typegraph.ResourceDefinition) error {
	file, err := generator.Generate(def)
	if err != nil {
		return err
	}
	for _, w := range file.Warnings {
		fmt.Fprintf(os.Stderr, "Schema validation for %s: %s\n", def.Name, w)
	}

	// Write to file (path includes the service folder, e.g. storage/storage_account_gen.go).
	outPath := filepath.Join("internal", "native", "services", file.RelPath)
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("creating output dir for %s: %w", def.Name, err)
	}
	if err := os.WriteFile(outPath, []byte(file.Source), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", outPath, err)
	}

	fmt.Printf("Generated %s (%d bytes)\n", outPath, len(file.Source))
	return nil
}
