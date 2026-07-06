package typegraph

// isFullyComputed reports whether an object is entirely server-populated.
func IsFullyComputed(typ *Type) bool {
	if typ.Kind != KindObject || len(typ.Properties) == 0 {
		return false
	}
	for _, prop := range typ.Properties {
		if prop.Flags.IsSystemManaged() || prop.Flags.IsReadOnly() {
			continue
		}
		if prop.Type.Kind == KindObject && IsFullyComputed(prop.Type) {
			continue
		}
		return false
	}
	return true
}

// EffectiveComputed folds explicit read-only, azwise-computed and fully read-only
// nested shapes into Computed-only schema attrs. It is model classification —
// read by both post-processing (typegraph) and the emitter (generator).
func EffectiveComputed(prop *Property) bool {
	if prop.ForceComputed || prop.Flags.IsReadOnly() {
		return true
	}
	if prop.Type.Kind == KindObject && IsFullyComputed(prop.Type) {
		return true
	}
	if prop.Type.Kind == KindArray && prop.Type.ElementType != nil &&
		prop.Type.ElementType.Kind == KindObject && IsFullyComputed(prop.Type.ElementType) {
		return true
	}
	return false
}
