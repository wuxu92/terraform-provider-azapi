package typegraph

// IsFullyComputed reports whether an object is entirely server-populated.
//
// It is only meaningful for a KindObject. A KindDiscriminated block is
// user-driven (the practitioner selects a variant), so it is never fully
// computed: the kind guard below returns false for it, and a discriminated
// property therefore keeps its enclosing object from being fully computed too.
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
	// A discriminated block is user-driven (the practitioner picks a variant), so
	// it is never wholly server-computed; leave it settable.
	return false
}
