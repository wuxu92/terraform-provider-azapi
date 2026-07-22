package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementStandaloneGateway provides resource knowledge for
// Microsoft.ApiManagement/gateways (the standalone API Management gateway,
// a top-level resource — not a child of service).
//
// Mirrors azurerm_api_management_standalone_gateway. sku (name+capacity) is
// required. backend_subnet_id and virtual_network_type are ForceNew and
// mutually required. tags is ForceNew.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_standalone_gateway_resource.go
//     Arguments (lines 54-116); Create body (lines 153-168); Timeout 30m.
//   - Microsoft.ApiManagement/gateways@2024-05-01 apigateway.ApiManagementGatewayResource:
//     sku.name (enum) + sku.capacity; properties.virtualNetworkType (enum),
//     properties.backend.subnet.id. provisioningState/targetProvisioningState/
//     createdAtUtc are read-only.
type ApiManagementStandaloneGateway struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementStandaloneGateway)(nil)

// NewApiManagementStandaloneGateway returns knowledge for the standalone gateway resource.
func NewApiManagementStandaloneGateway() *ApiManagementStandaloneGateway {
	return &ApiManagementStandaloneGateway{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/gateways",
			ApiVersions:  []string{"2024-05-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.backend.subnet.id"},
				{PropertyPath: "properties.virtualNetworkType"},
				// tags is ForceNew (commonschema.TagsForceNew()).
				{PropertyPath: "tags"},
			},
			RequiredFields: []string{
				"sku.name",
			},
			DefaultValues: []azwise.DefaultValue{
				// sku.capacity default 1 (schema line 87).
				{PropertyPath: "sku.capacity", Value: float64(1)},
				// virtual_network_type defaults to "None" in Create (line 148).
				{PropertyPath: "properties.virtualNetworkType", Value: "None"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Standard", "WorkspaceGatewayPremium", "WorkspaceGatewayStandard"},
					Message:       "sku.name must be one of Standard, WorkspaceGatewayPremium, WorkspaceGatewayStandard",
				},
				{
					PropertyPath:  "properties.virtualNetworkType",
					AllowedValues: []string{"External", "Internal", "None"},
					Message:       "virtual_network_type must be one of External, Internal, None",
				},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "sku.capacity", MinValue: azwise.Ptr(int64(1)), Message: "sku.capacity must be at least 1"},
			},
			// RequiredWith: backend_subnet_id <-> virtual_network_type are mutually
			// required (schema lines 100, 114). Both are ForceNew above.
			RequiredWith: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.backend.subnet.id", "properties.virtualNetworkType"},
					Message: "backend_subnet_id and virtual_network_type must be set together",
				},
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.targetProvisioningState",
				"properties.createdAtUtc",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementStandaloneGateway()) }
