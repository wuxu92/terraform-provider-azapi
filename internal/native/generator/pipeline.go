package generator

import (
	"fmt"

	"github.com/Azure/terraform-provider-azapi/internal/native/generator/customizers"
	"github.com/Azure/terraform-provider-azapi/internal/native/typegraph"
)

// GeneratedFile is one resource's emitted schema, ready to write. RelPath is the
// path under internal/native/services (e.g. "storage/storage_account_gen.go");
// Warnings holds non-fatal validation notes (bicep properties not yet covered by
// the emitted schema). The caller owns target selection and file I/O.
type GeneratedFile struct {
	RelPath  string
	Source   string
	Warnings []string
}

// BuildForGeneration builds the resource graph as it is compiled into the
// generated _gen.go: the runtime graph (parse + post-process) plus the
// per-resource customizers. This is the graph the compiled-schema verifier
// checks against; the runtime uses typegraph.BuildRuntimeGraph, which omits the
// customizer step (see ADR-0007).
func BuildForGeneration(data []byte) ([]*typegraph.ResourceDefinition, error) {
	defs, err := typegraph.BuildRuntimeGraph(data)
	if err != nil {
		return nil, err
	}
	customizers.Apply(defs)
	return defs, nil
}

// Generate runs the full ordered generation pipeline for one resource:
// flag-invariant check → emit → self-verify (emitted source vs the bicep type
// graph). Fatal problems (invariant violations, emit errors, extra/type
// mismatches) are returned as an error; non-fatal "missing in schema" notes are
// returned as Warnings on the file. This is the single home for the stage
// ordering that was formerly re-spelled at each call site (see ADR-0007).
func Generate(def *typegraph.ResourceDefinition) (GeneratedFile, error) {
	if vs := CheckFlagInvariants(def); len(vs) > 0 {
		return GeneratedFile{}, fmt.Errorf("schema invariant violations for %s:\n%s", def.Name, FormatViolations(vs))
	}

	source, err := EmitSchema(def)
	if err != nil {
		return GeneratedFile{}, fmt.Errorf("emitting %s: %w", def.Name, err)
	}

	// Verify the emitted schema covers every bicep body property and vice versa,
	// excluding the synthesized envelope attributes (name / parent reference / id).
	mismatches := ValidateEmittedSchema(source, def.Body, typegraph.EnvelopeAttrNames(def)...)
	var warnings []string
	fatal := 0
	for _, m := range mismatches {
		// extra and type_mismatch are fatal; missing is a warning.
		if m.Kind == typegraph.MismatchMissingInSchema {
			warnings = append(warnings, m.Detail)
		} else {
			fatal++
		}
	}
	if fatal > 0 {
		return GeneratedFile{}, fmt.Errorf("schema validation failed for %s:\n%s", def.Name, typegraph.FormatMismatches(mismatches))
	}

	return GeneratedFile{RelPath: FileName(def), Source: source, Warnings: warnings}, nil
}
