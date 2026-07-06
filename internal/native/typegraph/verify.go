package typegraph

import (
	"fmt"
	"strings"

	"github.com/Azure/terraform-provider-azapi/internal/native/naming"
)

// This file is the single schema-verification core (ADR-0008). Two adapters ask
// the same question — does a schema representation faithfully cover the bicep
// type graph, and vice versa? — against two substrates: the emitted Go source
// string (generator.ValidateEmittedSchema, a pre-compile codegen gate) and the
// compiled schema.Schema (validate.SchemaAgainstBicep, a post-init type check).
// They share this model and leaf rule; neither redefines it.

// Mismatch describes a discrepancy between a schema representation and the bicep
// type graph.
type Mismatch struct {
	Path   string // Dot-separated Terraform attribute path
	ARM    string // ARM property path (if known)
	Kind   MismatchKind
	Detail string
}

// MismatchKind identifies the type of schema-bicep mismatch.
type MismatchKind int

const (
	// MismatchExtraInSchema means the schema has an attribute not present in the
	// bicep type graph.
	MismatchExtraInSchema MismatchKind = iota

	// MismatchMissingInSchema means the bicep type graph has a property not
	// present in the schema.
	MismatchMissingInSchema

	// MismatchTypeMismatch means both sides have the property but the types don't
	// correspond.
	MismatchTypeMismatch
)

func (k MismatchKind) String() string {
	switch k {
	case MismatchExtraInSchema:
		return "extra_in_schema"
	case MismatchMissingInSchema:
		return "missing_in_schema"
	case MismatchTypeMismatch:
		return "type_mismatch"
	default:
		return "unknown"
	}
}

// FormatMismatches returns a human-readable summary of mismatches.
func FormatMismatches(mismatches []Mismatch) string {
	if len(mismatches) == 0 {
		return "No mismatches found."
	}

	var b strings.Builder
	extra, missing, typeMismatch := 0, 0, 0
	for _, m := range mismatches {
		switch m.Kind {
		case MismatchExtraInSchema:
			extra++
		case MismatchMissingInSchema:
			missing++
		case MismatchTypeMismatch:
			typeMismatch++
		}
	}
	b.WriteString(fmt.Sprintf("%d mismatches: %d extra in schema, %d missing in schema, %d type mismatches\n",
		len(mismatches), extra, missing, typeMismatch))
	for _, m := range mismatches {
		b.WriteString(fmt.Sprintf("  [%s] %s: %s\n", m.Kind, m.Path, m.Detail))
	}
	return b.String()
}

// JoinPath joins a dot-separated attribute path segment onto a prefix.
func JoinPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

// SchemaAttrName maps a bicep property (keyed by its ARM name) to the Terraform
// attribute name the emitter produces, and reports whether the property is
// skipped entirely. SystemManaged properties never appear in the schema. Both
// verification adapters call this so their traversals skip and rename identically.
func SchemaAttrName(armName string, prop *Property) (snake string, skip bool) {
	if prop.Flags.IsSystemManaged() {
		return "", true
	}
	return naming.CamelToSnake(armName), false
}

// RecurseChild reports the child object Type that schema verification descends
// into for a property's type: an object body, an array's object element, or a
// map's object element each nest a child attribute schema. It returns nil for
// leaf/scalar types (nothing to recurse into).
func RecurseChild(t *Type) *Type {
	if t == nil {
		return nil
	}
	switch t.Kind {
	case KindObject:
		return t
	case KindArray, KindMap:
		if t.ElementType != nil && t.ElementType.Kind == KindObject {
			return t.ElementType
		}
	}
	return nil
}
