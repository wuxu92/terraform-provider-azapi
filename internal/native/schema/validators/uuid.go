package validators

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type uuidValidator struct{}

func (uuidValidator) Description(context.Context) string {
	return "must be a valid UUID (e.g. 00000000-0000-0000-0000-000000000000)"
}

func (v uuidValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (uuidValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	str := req.ConfigValue
	if str.IsUnknown() || str.IsNull() {
		return
	}
	if !uuidPattern.MatchString(str.ValueString()) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid UUID",
			fmt.Sprintf("must be a valid UUID but got %q", str.ValueString()),
		)
	}
}

// UUID returns a string validator that requires a canonical UUID, ported from
// AzureRM's validation.IsUUID. Generic and cross-resource; reuse it before adding
// a new validator.
func UUID() validator.String {
	return uuidValidator{}
}
