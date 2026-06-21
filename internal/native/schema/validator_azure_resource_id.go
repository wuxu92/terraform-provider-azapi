package schema

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type azureResourceIDValidator struct{}

func (azureResourceIDValidator) Description(context.Context) string {
	return "must be a valid Azure resource ID (e.g. /subscriptions/.../resourceGroups/.../providers/...)"
}

func (v azureResourceIDValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (azureResourceIDValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	str := req.ConfigValue
	if str.IsUnknown() || str.IsNull() {
		return
	}
	if _, err := arm.ParseResourceID(str.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Azure resource ID",
			fmt.Sprintf("must be a valid Azure resource ID but got %q: %s", str.ValueString(), err),
		)
	}
}

// AzureResourceID returns a string validator that requires a parseable ARM
// resource ID, ported from AzureRM's azure.ValidateResourceID.
func AzureResourceID() validator.String {
	return azureResourceIDValidator{}
}
