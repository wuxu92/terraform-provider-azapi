package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ExpressRouteCircuit provides resource knowledge for
// Microsoft.Network/expressRouteCircuits.
//
// Mirrors azurerm_express_route_circuit. name and resource_group_name live on the
// operational envelope; only body/envelope-path knowledge is encoded here.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/express_route_circuit_resource.go
//     (schema resourceExpressRouteCircuit, expandExpressRouteCircuitSku, CRUD timeouts,
//     CustomizeDiff ForceNewIfChange on bandwidth_in_mbps lines 43-48)
//   - go-azure-sdk resource-manager/network/2025-01-01/expressroutecircuits:
//     id_expressroutecircuit.go (ARM type casing "expressRouteCircuits"),
//     model_expressroutecircuitpropertiesformat.go, model_expressroutecircuitsku.go,
//     model_expressroutecircuitserviceproviderproperties.go, constants.go (enum values)
//
// Not encoded (deliberate):
//   - service_key / service_provider_provisioning_state are server-computed read-only
//     fields (Computed in AzureRM, never written) → listed in ComputedFields.
//   - The mutual exclusion between the service-provider config
//     (serviceProviderProperties + bandwidthInMbps) and the port-based config
//     (expressRoutePort + bandwidthInGbps) is expressed via ConflictsWith/RequiredWith.
type ExpressRouteCircuit struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ExpressRouteCircuit)(nil)

// NewExpressRouteCircuit returns knowledge for the expressRouteCircuits resource.
func NewExpressRouteCircuit() *ExpressRouteCircuit {
	return &ExpressRouteCircuit{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/expressRouteCircuits",
			ApiVersions:  []string{"2025-01-01"},
			// location (commonschema.Location) replaces the circuit on change.
			// service_provider_name, peering_location and express_route_port_id are
			// ForceNew in AzureRM. bandwidth_in_mbps is conditionally ForceNew (only on
			// reduction) — handled in CheckForceNew below, not declaratively.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "properties.serviceProviderProperties.serviceProviderName"},
				{PropertyPath: "properties.serviceProviderProperties.peeringLocation"},
				{PropertyPath: "properties.expressRoutePort.id"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "sku.tier",
					AllowedValues: []string{"Basic", "Local", "Standard", "Premium"},
					Message:       "must be Basic, Local, Standard or Premium",
				},
				{
					PropertyPath:  "sku.family",
					AllowedValues: []string{"MeteredData", "UnlimitedData"},
					Message:       "must be MeteredData or UnlimitedData",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// allow_classic_operations Default:false.
				{PropertyPath: "properties.allowClassicOperations", Value: false},
				// rate_limiting_enabled Default:false -> enableDirectPortRateLimit.
				{PropertyPath: "properties.enableDirectPortRateLimit", Value: false},
			},
			// sku.name is hardcoded by AzureRM as "<tier>_<family>"; sku.tier and
			// sku.family are Required. All three must be present in the body.
			RequiredFields: []string{
				"sku.name",
				"sku.tier",
				"sku.family",
			},
			SensitiveFields: []string{
				"properties.authorizationKey",
				"properties.serviceKey",
			},
			ComputedFields: []string{
				"properties.serviceKey",
				"properties.serviceProviderProvisioningState",
			},
			// service_provider_name RequiredWith bandwidth_in_mbps + peering_location.
			// bandwidth_in_gbps RequiredWith express_route_port_id.
			RequiredWith: []azwise.RelationalRule{
				{
					Paths: []string{
						"properties.serviceProviderProperties.serviceProviderName",
						"properties.serviceProviderProperties.bandwidthInMbps",
						"properties.serviceProviderProperties.peeringLocation",
					},
					Message: "service_provider_name requires bandwidth_in_mbps and peering_location",
				},
				{
					Paths: []string{
						"properties.bandwidthInGbps",
						"properties.expressRoutePort.id",
					},
					Message: "bandwidth_in_gbps requires express_route_port_id",
				},
			},
			// The service-provider config and the port-based config are mutually
			// exclusive (AzureRM ConflictsWith).
			ConflictsWith: []azwise.RelationalRule{
				{
					Paths: []string{
						"properties.serviceProviderProperties.serviceProviderName",
						"properties.bandwidthInGbps",
						"properties.expressRoutePort.id",
					},
					Message: "service_provider_name conflicts with bandwidth_in_gbps and express_route_port_id",
				},
			},
		},
	}
}

// CheckForceNew extends the declarative check with the conditional bandwidth-shrink
// rule: reducing properties.serviceProviderProperties.bandwidthInMbps forces
// replacement, while raising it does not (mirrors the AzureRM CustomizeDiff
// ForceNewIfChange in express_route_circuit_resource.go lines 43-48).
func (s *ExpressRouteCircuit) CheckForceNew(oldBody, newBody map[string]interface{}) bool {
	if s.BaseKnowledge.CheckForceNew(oldBody, newBody) {
		return true
	}

	oldBw, oldOK := bandwidthInMbps(oldBody)
	newBw, newOK := bandwidthInMbps(newBody)
	if oldOK && newOK && newBw < oldBw {
		return true
	}

	return false
}

// bandwidthInMbps extracts properties.serviceProviderProperties.bandwidthInMbps.
func bandwidthInMbps(body map[string]interface{}) (float64, bool) {
	props, ok := body["properties"].(map[string]interface{})
	if !ok {
		return 0, false
	}
	sp, ok := props["serviceProviderProperties"].(map[string]interface{})
	if !ok {
		return 0, false
	}
	v, ok := azwise.ToFloat64(sp["bandwidthInMbps"])
	return v, ok
}

func init() { azwise.Register(NewExpressRouteCircuit()) }
