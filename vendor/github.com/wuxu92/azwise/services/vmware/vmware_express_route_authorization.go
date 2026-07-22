package vmware

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VmwareExpressRouteAuthorization provides resource knowledge for
// Microsoft.AVS/privateClouds/authorizations.
//
// Mirrors azurerm_vmware_express_route_authorization. The resource sends an empty
// ExpressRouteAuthorization body on create (all properties are server-generated);
// there are no settable ARM body fields.
//
// Sources:
//   - internal/services/vmware/vmware_express_route_authorization_resource.go
//     (schema 40-65: name ForceNew StringIsNotEmpty; private_cloud_id ForceNew (parent
//     ref, not a body path); express_route_authorization_id Computed;
//     express_route_authorization_key Computed Sensitive; timeouts Create/Delete 30m
//     Read 5m; create 95 → ExpressRouteAuthorization{} (empty body)).
//   - go-azure-sdk resource-manager/vmware/2022-05-01/authorizations:
//     model_expressrouteauthorizationproperties.go (expressRouteAuthorizationId/
//     expressRouteAuthorizationKey/expressRouteId/provisioningState all read-only),
//     id_authorization.go (segment casing "Microsoft.AVS"/"privateClouds"/"authorizations").
type VmwareExpressRouteAuthorization struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VmwareExpressRouteAuthorization)(nil)

// NewVmwareExpressRouteAuthorization returns knowledge for the
// privateClouds/authorizations resource.
func NewVmwareExpressRouteAuthorization() *VmwareExpressRouteAuthorization {
	return &VmwareExpressRouteAuthorization{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AVS/privateClouds/authorizations",
			ApiVersions:  []string{"2022-05-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			StringRules: []azwise.StringRule{
				{
					// name: StringIsNotEmpty.
					MinLength: 1,
					Message:   "must not be empty",
				},
			},
			ComputedFields: []string{
				"properties.expressRouteAuthorizationId",
				"properties.expressRouteAuthorizationKey",
				"properties.expressRouteId",
				"properties.provisioningState",
			},
			// NOTE: private_cloud_id is the parent private-cloud reference (ForceNew in
			// AzureRM) and has no ARM body path — it is encoded into the resource ID.
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewVmwareExpressRouteAuthorization()) }
