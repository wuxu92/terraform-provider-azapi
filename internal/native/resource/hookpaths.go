package resource

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/terraform-provider-azapi/internal/native/services"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ValidateRelationalPaths checks that every attribute path declared in the named
// resource's Hooks.Relational resolves against that resource's schema. Hook paths are
// plain strings the Go compiler cannot relate to the generated schema, so a typo
// (e.g. "propertis.foo") would otherwise surface only when a practitioner validates a
// config for that resource. Call it for every registered resource from a test to catch
// such typos in CI. Returns one error per unresolved path; nil when all resolve.
func ValidateRelationalPaths(ctx context.Context, name string) []error {
	d, ok := services.Registry[name]
	if !ok {
		return []error{fmt.Errorf("native: no descriptor registered for %q", name)}
	}
	h := hookRegistry[name]
	if h == nil || len(h.Relational) == 0 {
		return nil
	}
	s := d.Schema()
	var errs []error
	for _, rc := range h.Relational {
		for _, p := range rc.Paths {
			if err := schemaHasPath(ctx, s, p); err != nil {
				errs = append(errs, fmt.Errorf("%s: relational path %q does not resolve in schema: %w", name, p, err))
			}
		}
	}
	return errs
}

// schemaHasPath reports whether a dot-separated snake_case attribute path resolves to
// a type in the schema, mirroring how the relational validator walks it at runtime
// (AttributeName steps only; relational paths never index into list/set elements).
func schemaHasPath(ctx context.Context, s schema.Schema, dotted string) error {
	segs := strings.Split(dotted, ".")
	steps := make([]tftypes.AttributePathStep, len(segs))
	for i, seg := range segs {
		steps[i] = tftypes.AttributeName(seg)
	}
	_, err := s.TypeAtTerraformPath(ctx, tftypes.NewAttributePathWithSteps(steps))
	return err
}
