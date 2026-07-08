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
	"github.com/Azure/terraform-provider-azapi/internal/native/typegraph"
	"github.com/Azure/terraform-provider-azapi/internal/native/validate"

	// generated provides the Registry/Descriptor types; the all aggregator's blank
	// import runs each service package's init() to populate services.Registry.
	"github.com/Azure/terraform-provider-azapi/internal/native/services"
	_ "github.com/Azure/terraform-provider-azapi/internal/native/services/all"
)

func main() {
	res := flag.String("r", "", "resource name to validate (e.g. azapi_storage_account); omit to validate all")
	flag.StringVar(res, "res", "", "resource name to validate (e.g. azapi_storage_account); omit to validate all")
	flag.Parse()

	// Determine which schemas to validate
	targets := services.Registry
	if *res != "" {
		d, ok := services.Registry[*res]
		if !ok {
			fmt.Fprintf(os.Stderr, "Unknown resource: %s\nAvailable:\n", *res)
			for k := range services.Registry {
				fmt.Fprintf(os.Stderr, "  %s\n", k)
			}
			os.Exit(1)
		}
		targets = map[string]services.Descriptor{*res: d}
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

		// Build the graph exactly as the generated schema was compiled from
		// (post-process + customizers), then select the resource under test.
		defs, err := generator.BuildForGeneration(typesData)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR %s: building generation graph: %v\n", name, err)
			totalErrors++
			continue
		}
		var pdef *typegraph.ResourceDefinition
		for _, def := range defs {
			if def.Name == tag {
				pdef = def
				break
			}
		}
		if pdef == nil {
			fmt.Fprintf(os.Stderr, "ERROR %s: resource %s not found in %s\n", name, tag, location)
			totalErrors++
			continue
		}
		body := pdef.Body

		// Validate the body; exclude the synthesized envelope attributes (name /
		// parent reference / id) plus any behavior-only Meta attributes a customizer
		// attached (e.g. purge_on_destroy) — none are part of the bicep body. This is
		// the same exclusion set the generation pipeline uses (EnvelopeAttrNames).
		mismatches := validate.SchemaAgainstBicep(s, body, typegraph.EnvelopeAttrNames(pdef)...)

		// Flag-invariant restraints: Default ⟹ Optional+Computed (not Required /
		// read-only) and Default ∈ its own validators.
		inv := generator.CheckFlagInvariants(pdef)

		errors := 0
		warnings := 0
		for _, m := range mismatches {
			switch m.Kind {
			case typegraph.MismatchExtraInSchema, typegraph.MismatchTypeMismatch:
				errors++
			case typegraph.MismatchMissingInSchema:
				warnings++
			}
		}

		errors += len(inv)
		for _, v := range inv {
			fmt.Fprintf(os.Stderr, "INVARIANT %s (%s): %s\n", name, tag, v.String())
		}
		switch {
		case len(mismatches) > 0:
			fmt.Fprintf(os.Stderr, "FAIL %s (%s): %s", name, tag, typegraph.FormatMismatches(mismatches))
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
