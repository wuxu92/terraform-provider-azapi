package analysisservices

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AnalysisServicesServer provides resource knowledge for
// Microsoft.AnalysisServices/servers.
//
// Mirrors azurerm_analysis_services_server. name/resource_group_name/location are
// ForceNew in AzureRM but are operational-envelope fields, so they are not encoded
// as ARM-body ForceNew rules here.
//
// sku maps to the top-level `sku.name` (ResourceSku is a sibling of `properties`,
// not nested under it). admin_users maps to properties.asAdministrators.members;
// power_bi_service_enabled and ipv4_firewall_rule are folded into
// properties.ipV4FirewallSettings.{enablePowerBIService,firewallRules}. The server
// name validator (validate.ServerName) is an envelope-name check handled at the
// schema/customizer layer, and the per-rule IPv4 range validators
// (range_start/range_end) target array elements, so neither is expressed as a
// declarative StringRule here.
//
// Sources:
//   - terraform-provider-azurerm internal/services/analysisservices/analysis_services_server_resource.go
//     schema + Create/Update (sku Required with fixed SKU list; querypool default
//     "All"; backup_blob_container_uri Sensitive; timeouts 30m/5m/30m/30m)
//   - go-azure-sdk resource-manager/analysisservices/2017-08-01/servers
//     AnalysisServicesServer{Sku required} + AnalysisServicesServerProperties
//     (asAdministrators/backupBlobContainerUri/ipV4FirewallSettings/
//     querypoolConnectionMode settable; serverFullName/provisioningState/state
//     read-only)
type AnalysisServicesServer struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AnalysisServicesServer)(nil)

// NewAnalysisServicesServer returns knowledge for the servers resource.
func NewAnalysisServicesServer() *AnalysisServicesServer {
	return &AnalysisServicesServer{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AnalysisServices/servers",
			ApiVersions:  []string{"2017-08-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"sku.name",
			},
			// AzureRM defaults querypool_connection_mode to "All".
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.querypoolConnectionMode", Value: "All"},
			},
			StringRules: []azwise.StringRule{
				// AzureRM's fixed SKU list; ResourceSku.Name is a plain string in the
				// SDK (no enum), so the AzureRM allow-list is authoritative.
				{PropertyPath: "sku.name", AllowedValues: []string{
					"D1", "B1", "B2", "S0", "S1", "S2", "S4", "S8", "S9", "S8v2", "S9v2",
				}, Message: "sku must be a valid Analysis Services SKU"},
				// Full ARM SDK ConnectionMode enum.
				{PropertyPath: "properties.querypoolConnectionMode", AllowedValues: []string{
					"All", "ReadOnly",
				}, Message: "querypool connection mode must be All or ReadOnly"},
				{PropertyPath: "properties.backupBlobContainerUri", MinLength: 1, Message: "backup blob container uri must not be empty"},
			},
			SensitiveFields: []string{
				"properties.backupBlobContainerUri",
			},
			ComputedFields: []string{
				"properties.serverFullName",
				"properties.provisioningState",
				"properties.state",
			},
		},
	}
}

func init() { azwise.Register(NewAnalysisServicesServer()) }
