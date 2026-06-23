// Command azapin-validate validates compiled native schemas against their
// source bicep type definitions. Each schema self-describes its resource type
// via an [azapin:ResourceType@Version] tag in the Description field.
//
// Usage:
//
//	go run ./internal/native/cmd/azapin-validate/                       # validate all
//	go run ./internal/native/cmd/azapin-validate/ -r azapi_storage_account  # validate one

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Azure/terraform-provider-azapi/internal/azure"
	"github.com/Azure/terraform-provider-azapi/internal/native/generator"
	"github.com/Azure/terraform-provider-azapi/internal/native/generator/customizers"
	"github.com/Azure/terraform-provider-azapi/internal/native/validate"

	// generated provides the Registry/Descriptor types; the all aggregator's blank
	// import runs each service package's init() to populate generated.Registry.
	"github.com/Azure/terraform-provider-azapi/internal/native/generated"
	_ "github.com/Azure/terraform-provider-azapi/internal/native/generated/all"
)

func main() {
	res := flag.String("r", "", "resource name to validate (e.g. azapi_storage_account); omit to validate all")
	flag.StringVar(res, "res", "", "resource name to validate (e.g. azapi_storage_account); omit to validate all")
	flag.Parse()

	// Determine which schemas to validate
	targets := generated.Registry
	if *res != "" {
		d, ok := generated.Registry[*res]
		if !ok {
			fmt.Fprintf(os.Stderr, "Unknown resource: %s\nAvailable:\n", *res)
			for k := range generated.Registry {
				fmt.Fprintf(os.Stderr, "  %s\n", k)
			}
			os.Exit(1)
		}
		targets = map[string]generated.Descriptor{*res: d}
	}

	totalErrors := 0
	totalWarnings := 0
	validated := 0

	for name, d := range targets {
		s := d.Schema()

		// Resource tag comes straight from the descriptor.
		tag := d.ARMType + "@" + d.APIVersion

		// Resolve the bicep types.json that defines this resource from the azure
		// schema loader (the same embedded source the provider runtime loads).
		location, err := azure.GetResourceTypeLocation(d.ARMType, d.APIVersion)
		if err != nil {
			fmt.Fprintf(os.Stderr, "SKIP %s: %v\n", name, err)
			continue
		}

		typesData, err := azure.StaticFiles.ReadFile("generated/" + location)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR %s: reading %s: %v\n", name, location, err)
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
			fmt.Fprintf(os.Stderr, "ERROR %s: resource %s not found in %s\n", name, tag, location)
			totalErrors++
			continue
		}

		// Apply the same generation pipeline: post-process, then customizers, so the
		// body graph matches what the generated schema was built from.
		pdef := &generator.ResourceDefinition{Name: tag, Body: body}
		generator.PostProcess([]*generator.ResourceDefinition{pdef})
		customizers.Apply([]*generator.ResourceDefinition{pdef})

		// Validate the body; exclude the synthesized envelope attributes.
		mismatches := validate.SchemaAgainstBicep(s, body, "name", d.ParentAttr, "id")

		// Flag-invariant restraints: Default ⟹ Optional+Computed (not Required /
		// read-only) and Default ∈ its own validators.
		inv := generator.CheckFlagInvariants(pdef)

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

		errors += len(inv)
		for _, v := range inv {
			fmt.Fprintf(os.Stderr, "INVARIANT %s (%s): %s\n", name, tag, v.String())
		}
		switch {
		case len(mismatches) > 0:
			fmt.Fprintf(os.Stderr, "FAIL %s (%s): %s", name, tag, validate.FormatMismatches(mismatches))
		case errors == 0 && warnings == 0:
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
