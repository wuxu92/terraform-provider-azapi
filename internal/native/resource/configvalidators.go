package resource

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/terraform-provider-azapi/internal/native/services"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var _ resource.ResourceWithConfigValidators = &Base{}

// ConfigValidators returns resource-level cross-property validators declared by the
// resource's hand-written hooks. Two sources, in order: (1) the declarative
// Hooks.Relational constraints (ConflictsWith / RequiredWith / ExactlyOneOf /
// AtLeastOneOf / AtMostOneOf), each lowered to a framework ConfigValidator, then
// (2) any prebuilt Hooks.ConfigValidator objects verbatim. Single-attribute
// validators (enum, range, length) stay in the generated schema; these span
// multiple attributes and so live at resource scope, analogous to the runtime-added
// timeouts block (see composeSchema). Both sources are hook-owned, not generated, so
// they can be tuned without regenerating the resource.
func (b *Base) ConfigValidators(context.Context) []resource.ConfigValidator {
	if b.hooks == nil {
		return nil
	}
	out := make([]resource.ConfigValidator, 0, len(b.hooks.Relational)+len(b.hooks.ConfigValidators))
	for _, rc := range b.hooks.Relational {
		out = append(out, relationalValidator{constraint: rc})
	}
	out = append(out, b.hooks.ConfigValidators...)
	if len(out) == 0 {
		return nil
	}
	return out
}

// relationalValidator enforces one resource-level cross-property constraint. For
// ConflictsWith and RequiredWith, constraint.Paths[0] is the subject.
type relationalValidator struct {
	constraint services.RelationalConstraint
}

var _ resource.ConfigValidator = relationalValidator{}

func (v relationalValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v relationalValidator) MarkdownDescription(context.Context) string {
	label := map[services.RelationalKind]string{
		services.ConflictsWith: "conflicting attributes",
		services.RequiredWith:  "attributes required together",
		services.ExactlyOneOf:  "exactly one of the attributes",
		services.AtLeastOneOf:  "at least one of the attributes",
		services.AtMostOneOf:   "at most one of the attributes",
	}[v.constraint.Kind]
	return fmt.Sprintf("Cross-property constraint (%s): %s", label, joinPaths(v.constraint.Paths))
}

func (v relationalValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if len(v.constraint.Paths) < 2 {
		return
	}
	set := make([]bool, len(v.constraint.Paths))
	for i, p := range v.constraint.Paths {
		isSet, unknown, diags := attrIsSet(ctx, req.Config, p)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}
		if unknown {
			return // delay validation until every involved value is known
		}
		set[i] = isSet
	}

	subject := subjectPath(v.constraint.Paths[0])
	switch v.constraint.Kind {
	case services.ConflictsWith:
		if !set[0] {
			return
		}
		for i := 1; i < len(set); i++ {
			if set[i] {
				resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(subject,
					"Invalid Attribute Combination",
					relationalDetail(v.constraint.Message,
						fmt.Sprintf("%q cannot be set together with %q.", v.constraint.Paths[0], v.constraint.Paths[i]))))
			}
		}
	case services.RequiredWith:
		if !set[0] {
			return
		}
		for i := 1; i < len(set); i++ {
			if !set[i] {
				resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(subject,
					"Invalid Attribute Combination",
					relationalDetail(v.constraint.Message,
						fmt.Sprintf("%q must be set when %q is set.", v.constraint.Paths[i], v.constraint.Paths[0]))))
			}
		}
	case services.ExactlyOneOf:
		count := countSet(set)
		if count != 1 {
			resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(subject,
				"Invalid Attribute Combination",
				relationalDetail(v.constraint.Message,
					fmt.Sprintf("Exactly one of [%s] must be set, but %d are set.", joinPaths(v.constraint.Paths), count))))
		}
	case services.AtLeastOneOf:
		if countSet(set) == 0 {
			resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(subject,
				"Invalid Attribute Combination",
				relationalDetail(v.constraint.Message,
					fmt.Sprintf("At least one of [%s] must be set.", joinPaths(v.constraint.Paths)))))
		}
	case services.AtMostOneOf:
		if count := countSet(set); count > 1 {
			resp.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(subject,
				"Invalid Attribute Combination",
				relationalDetail(v.constraint.Message,
					fmt.Sprintf("At most one of [%s] may be set, but %d are set.", joinPaths(v.constraint.Paths), count))))
		}
	}
}

// attrIsSet reports whether the config value at the given dot-separated snake_case
// path is present (non-null, known). unknown is true when the value is still
// unknown, signalling the caller to defer validation.
func attrIsSet(ctx context.Context, cfg tfsdk.Config, dotted string) (set, unknown bool, diags diag.Diagnostics) {
	if dotted == "" {
		return false, false, nil
	}
	var v attr.Value
	diags = cfg.GetAttribute(ctx, subjectPath(dotted), &v)
	if diags.HasError() || v == nil {
		return false, false, diags
	}
	if v.IsUnknown() {
		return false, true, diags
	}
	return !v.IsNull(), false, diags
}

// subjectPath converts a dot-separated snake_case path ("properties.foo.bar") into
// a framework attribute path. Attribute names never contain a dot, so splitting on
// "." recovers the segments exactly.
func subjectPath(dotted string) path.Path {
	segs := strings.Split(dotted, ".")
	p := path.Root(segs[0])
	for _, s := range segs[1:] {
		p = p.AtName(s)
	}
	return p
}

func countSet(set []bool) int {
	n := 0
	for _, s := range set {
		if s {
			n++
		}
	}
	return n
}

func relationalDetail(message, fallback string) string {
	if strings.TrimSpace(message) != "" {
		return message
	}
	return fallback
}

func joinPaths(paths []string) string { return strings.Join(paths, ", ") }
