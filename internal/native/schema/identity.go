package schema

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const defaultManagedServiceIdentityDescription = "The identity of the resource."

// ManagedServiceIdentity returns the common ARM managed identity block used by
// native resources. The top-level principal_id / tenant_id are the SYSTEM-assigned
// identity's coordinates and are server-populated only when the identity type includes
// SystemAssigned; UseStateForSystemPrincipal holds them stable and pins null for a
// UserAssigned-only identity (which Azure never populates). The per-user-assigned
// identity client_id / principal_id are always server-populated, so they use
// UseNonNullStateForUnknown to avoid Terraform's "was null, but now ..." apply error
// when Azure fills those read-only values during the same operation.
func ManagedServiceIdentity(required bool, desc string) schema.SingleNestedAttribute {
	description := desc
	if description == "" {
		description = defaultManagedServiceIdentityDescription
	}

	typeAttribute := schema.StringAttribute{
		Description: "Type of managed service identity.",
		Validators: []validator.String{
			stringvalidator.OneOf(
				"None",
				"SystemAssigned",
				"UserAssigned",
				"SystemAssigned, UserAssigned",
			),
		},
	}
	if required {
		typeAttribute.Required = true
	} else {
		typeAttribute.Optional = true
		typeAttribute.Computed = true
		typeAttribute.PlanModifiers = []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		}
	}

	return schema.SingleNestedAttribute{
		Description: description,
		Optional:    true,
		Computed:    true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Attributes: map[string]schema.Attribute{
			"principal_id": schema.StringAttribute{
				Description: "Principal ID of the managed identity.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseStateForSystemPrincipal(),
				},
			},
			"tenant_id": schema.StringAttribute{
				Description: "Tenant ID of the managed identity.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseStateForSystemPrincipal(),
				},
			},
			"type": typeAttribute,
			"user_assigned_identities": schema.MapNestedAttribute{
				Description: "User-assigned identities keyed by ARM resource ID.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"client_id": schema.StringAttribute{
							Description: "Client ID of the user-assigned identity.",
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseNonNullStateForUnknown(),
							},
						},
						"principal_id": schema.StringAttribute{
							Description: "Principal ID of the user-assigned identity.",
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseNonNullStateForUnknown(),
							},
						},
					},
				},
			},
		},
	}
}
