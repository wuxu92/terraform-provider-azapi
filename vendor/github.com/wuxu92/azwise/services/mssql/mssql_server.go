package mssql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MsSqlServer provides resource knowledge for Microsoft.Sql/servers
// (azurerm_mssql_server).
//
// Sources:
//   - AzureRM internal/services/mssql/mssql_server_resource.go
//     :54-59   (timeouts: create/update/delete 60m, read 5m)
//     :62-67   (name ForceNew, validate.ValidateMsSqlServerName)
//     :73-81   (version Required ForceNew, StringInSlice "2.0"/"12.0")
//     :83-90   (administrator_login Optional+Computed ForceNew)
//     :92-98   (administrator_login_password Sensitive)
//     :183-196 (minimum_tls_version default 1.2, public_network_access default true,
//     outbound_network_restriction default false)
//     :267-336 (create mapping: properties.version/administratorLogin/minimalTlsVersion/
//     publicNetworkAccess/restrictOutboundNetworkAccess/keyId/primaryUserAssignedIdentityId)
//   - AzureRM internal/services/mssql/validate/mssql.go:13-19 (name regex)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/servers:
//     model_serverproperties.go, constants.go (PossibleValuesForMinimalTlsVersion)
//
// NOTE: connection_policy maps to the separate Microsoft.Sql/servers/connectionPolicies
// sub-resource, and express_vulnerability_assessment_enabled maps to
// Microsoft.Sql/servers/sqlVulnerabilityAssessments — neither is part of the servers
// body, so they are not represented here.
type MsSqlServer struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MsSqlServer)(nil)

// NewMsSqlServer returns knowledge for the Microsoft.Sql/servers resource.
func NewMsSqlServer() *MsSqlServer {
	return &MsSqlServer{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/servers",
			ApiVersions:  []string{"2023-08-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.version"},
				{PropertyPath: "properties.administratorLogin"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[0-9a-z]([-0-9a-z]{0,61}[0-9a-z])?$`,
					Message:      "name can contain only lowercase letters, numbers and '-', can't start or end with '-', and can't be longer than 63 characters",
				},
				{
					PropertyPath:  "properties.version",
					AllowedValues: []string{"2.0", "12.0"},
					Message:       "version must be one of 2.0 or 12.0",
				},
				{
					// AzureRM restricts this to "1.2" (or 1.0/1.1/1.2/Disabled pre-5.0),
					// but the ARM API accepts the full MinimalTlsVersion enum set.
					PropertyPath:  "properties.minimalTlsVersion",
					AllowedValues: []string{"None", "1.0", "1.1", "1.2", "1.3"},
					Message:       "minimal_tls_version must be one of None, 1.0, 1.1, 1.2 or 1.3",
				},
			},
			SensitiveFields: []string{
				"properties.administratorLoginPassword",
			},
			RequiredFields: []string{
				"properties.version",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.minimalTlsVersion", Value: "1.2"},
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				{PropertyPath: "properties.restrictOutboundNetworkAccess", Value: "Disabled"},
			},
			ComputedFields: []string{
				"properties.fullyQualifiedDomainName",
				"properties.state",
				"properties.externalGovernanceStatus",
				"properties.privateEndpointConnections",
			},
		},
	}
}

func init() { azwise.Register(NewMsSqlServer()) }
