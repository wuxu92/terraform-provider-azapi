package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SubnetServiceEndpointStoragePolicy provides resource knowledge for
// Microsoft.Network/serviceEndpointPolicies.
//
// Mirrors azurerm_subnet_service_endpoint_storage_policy (the AzureRM name is
// storage-scoped, but the ARM resource is the generic serviceEndpointPolicies
// type). name and resource_group are envelope-owned; name is ForceNew; location
// forces replacement.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/subnet_service_endpoint_storage_policy_resource.go
//     (schema, expandServiceEndpointPolicyDefinitions, CRUD timeouts 30/5/30/30)
//   - terraform-provider-azurerm internal/services/network/validate/subnet_service_endpoint_storage_policy_name.go (name regex)
//   - go-azure-sdk resource-manager/network/2025-01-01/serviceendpointpolicies:
//     model_serviceendpointpolicypropertiesformat.go, id_serviceendpointpolicy.go
//     (segment casing "serviceEndpointPolicies")
//
// Not encoded (deliberate):
//   - the definition block folds into
//     properties.serviceEndpointPolicyDefinitions[*]; its per-definition fields
//     (service enum Microsoft.Storage/Global, service_resources IDs, description
//     StringLenBetween(0,140)) live under an array element which azwise/azapin
//     cannot lower through "[*]". Skipped, not emitted.
type SubnetServiceEndpointStoragePolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SubnetServiceEndpointStoragePolicy)(nil)

// NewSubnetServiceEndpointStoragePolicy returns knowledge for the
// serviceEndpointPolicies resource.
func NewSubnetServiceEndpointStoragePolicy() *SubnetServiceEndpointStoragePolicy {
	return &SubnetServiceEndpointStoragePolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/serviceEndpointPolicies",
			ApiVersions:  []string{"2025-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					Regex:   `^[^\W_]([\w.\-]{0,78}[\w])?$`,
					Message: "name can be up to 80 chars, must begin with an alphanumeric and end with an alphanumeric or underscore, and may contain alphanumerics, periods, hyphens or underscores",
				},
			},
		},
	}
}

func init() { azwise.Register(NewSubnetServiceEndpointStoragePolicy()) }
