package typegraph

// BuildRuntimeGraph parses a bicep types.json and applies post-processing
// (defaults, validators, the azwise overlay, and the operational-envelope spec),
// returning the resource definitions the provider runtime reads.
//
// It deliberately does NOT run the per-resource generation customizers: those
// mutate the graph for emission only (validators, plan modifiers, schema flags
// baked into the generated _gen.go) and must never appear in the runtime graph
// the mapper reads. Generation code builds on top of this via
// generator.BuildForGeneration, which adds customizers.Apply. Keeping the
// customizer step out of this function — and out of the typegraph package the
// runtime imports — makes the runtime/as-compiled divergence a compile-time
// impossibility (see ADR-0007).
func BuildRuntimeGraph(data []byte) ([]*ResourceDefinition, error) {
	defs, err := ParseTypesJSON(data)
	if err != nil {
		return nil, err
	}
	PostProcess(defs)
	return defs, nil
}
