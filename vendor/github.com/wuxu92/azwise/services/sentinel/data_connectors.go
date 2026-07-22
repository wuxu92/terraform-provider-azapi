// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package sentinel

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DataConnectors provides resource knowledge for
// Microsoft.SecurityInsights/dataConnectors (an extension resource on a
// Microsoft.OperationalInsights/workspaces scope).
//
// Contributing Terraform resources (all map to the dataConnectors ARM type,
// discriminated by the top-level `kind`):
//   - azurerm_sentinel_data_connector_aws_cloud_trail                          (AmazonWebServicesCloudTrail)
//   - azurerm_sentinel_data_connector_aws_s3                                   (AmazonWebServicesS3)
//   - azurerm_sentinel_data_connector_azure_active_directory                   (AzureActiveDirectory)
//   - azurerm_sentinel_data_connector_azure_advanced_threat_protection         (AzureAdvancedThreatProtection)
//   - azurerm_sentinel_data_connector_azure_security_center                    (AzureSecurityCenter)
//   - azurerm_sentinel_data_connector_dynamics_365                             (Dynamics365)
//   - azurerm_sentinel_data_connector_iot                                      (IOT)
//   - azurerm_sentinel_data_connector_microsoft_cloud_app_security             (MicrosoftCloudAppSecurity)
//   - azurerm_sentinel_data_connector_microsoft_defender_advanced_threat_protection (MicrosoftDefenderAdvancedThreatProtection)
//   - azurerm_sentinel_data_connector_microsoft_threat_intelligence            (MicrosoftThreatIntelligence)
//   - azurerm_sentinel_data_connector_microsoft_threat_protection              (MicrosoftThreatProtection)
//   - azurerm_sentinel_data_connector_office_365                               (Office365)
//   - azurerm_sentinel_data_connector_office_365_project                       (Office365Project)
//   - azurerm_sentinel_data_connector_office_atp                               (OfficeATP)
//   - azurerm_sentinel_data_connector_office_irm                               (OfficeIRM)
//   - azurerm_sentinel_data_connector_office_power_bi                          (OfficePowerBI)
//   - azurerm_sentinel_data_connector_threat_intelligence                      (ThreatIntelligence)
//   - azurerm_sentinel_data_connector_threat_intelligence_taxii                (ThreatIntelligenceTaxii)
//
// Merge policy: only universal knowledge is unioned. Per-kind value constraints
// (tenantId, awsRoleArn/roleArn) fire only when their path is present in the
// body, so they are safe across the discriminated union.
//
// Sources:
//   - terraform-provider-azurerm internal/services/sentinel/sentinel_data_connector.go
//     (assertDataConnectorKind L54-96)
//   - internal/services/sentinel/sentinel_data_connector_azure_active_directory_resource.go
//     (schema L35-57, body L93-104)
//   - internal/services/sentinel/sentinel_data_connector_aws_cloud_trail_resource.go (aws_role_arn L54)
//   - internal/services/sentinel/sentinel_data_connector_aws_s3_resource.go (aws_role_arn L51, sqs_urls)
//   - go-azure-sdk resource-manager/securityinsights/2022-10-01-preview/dataconnectors:
//     id_dataconnector.go (Microsoft.SecurityInsights/dataConnectors), constants.go
//     (DataConnectorKind), model_aaddataconnectorproperties.go (tenantId),
//     model_awscloudtraildataconnectorproperties.go (awsRoleArn),
//     model_awss3dataconnectorproperties.go (roleArn, sqsUrls)
type DataConnectors struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DataConnectors)(nil)

// NewDataConnectors returns knowledge for the dataConnectors resource type.
func NewDataConnectors() *DataConnectors {
	return &DataConnectors{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.SecurityInsights/dataConnectors",
			ApiVersions:  []string{"2022-10-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				// tenant_id is ForceNew for the tenant-scoped connector kinds
				// (AzureActiveDirectory, AzureAdvancedThreatProtection,
				// AzureSecurityCenter, Dynamics365, IOT, ...). Fires only when present.
				{PropertyPath: "properties.tenantId"},
			},
			// "kind" is the discriminator and is required for every dataConnectors body.
			RequiredFields: []string{"kind"},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					MinLength:    1,
					Message:      "data connector name must not be empty",
				},
				{
					PropertyPath: "kind",
					// Full ARM SDK DataConnectorKind set.
					AllowedValues: []string{
						"APIPolling", "AmazonWebServicesCloudTrail", "AmazonWebServicesS3",
						"AzureActiveDirectory", "AzureAdvancedThreatProtection", "AzureSecurityCenter",
						"Dynamics365", "GenericUI", "IOT", "MicrosoftCloudAppSecurity",
						"MicrosoftDefenderAdvancedThreatProtection", "MicrosoftThreatIntelligence",
						"MicrosoftThreatProtection", "OfficeATP", "OfficeIRM", "OfficePowerBI",
						"Office365", "Office365Project", "ThreatIntelligence", "ThreatIntelligenceTaxii",
					},
					Message: "kind must be a supported Sentinel data connector kind",
				},
				{
					// tenant_id (validation.IsUUID) for tenant-scoped kinds.
					PropertyPath: "properties.tenantId",
					Regex:        `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
					Message:      "tenant_id must be a valid UUID",
				},
				{
					// aws_role_arn on the AmazonWebServicesCloudTrail kind (validate.IsARN).
					PropertyPath: "properties.awsRoleArn",
					Regex:        `^arn:aws:iam::[0-9]{12}:role/.+$`,
					Message:      "aws_role_arn must be a valid AWS IAM role ARN",
				},
				{
					// aws_role_arn on the AmazonWebServicesS3 kind (validate.IsARN).
					PropertyPath: "properties.roleArn",
					Regex:        `^arn:aws:iam::[0-9]{12}:role/.+$`,
					Message:      "aws_role_arn must be a valid AWS IAM role ARN",
				},
			},
		},
	}
}

func init() { azwise.Register(NewDataConnectors()) }
