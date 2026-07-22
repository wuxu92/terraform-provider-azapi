package iothub

import (
	"time"

	"github.com/wuxu92/azwise"
)

// IotHubDpsCertificate provides resource knowledge for
// Microsoft.Devices/provisioningServices/certificates.
//
// Contributing Terraform resource: azurerm_iothub_dps_certificate.
//
// Sources:
//   - terraform-provider-azurerm internal/services/iothub/iothub_dps_certificate_resource.go
//     (schema L41-71, Create body L96-103, timeouts L34-39)
//   - terraform-provider-azurerm internal/services/iothub/validate/iot_hub_dps_certificate_name.go
//     (IoTHubDpsCertificateName L11-26)
//   - go-azure-sdk resource-manager/deviceprovisioningservices/2022-02-05/dpscertificate:
//     model_certificateproperties.go (Certificate, IsVerified)
//
// Notes:
//   - name/iot_dps_name/resource_group_name are envelope / parent-reference fields.
//   - certificate_content is Sensitive and maps to properties.certificate.
//   - is_verified is ForceNew in AzureRM and maps to properties.isVerified (default false).
type IotHubDpsCertificate struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*IotHubDpsCertificate)(nil)

// NewIotHubDpsCertificate returns knowledge for the provisioningServices/certificates resource.
func NewIotHubDpsCertificate() *IotHubDpsCertificate {
	return &IotHubDpsCertificate{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Devices/provisioningServices/certificates",
			ApiVersions:  []string{"2022-02-05"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.isVerified"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[0-9a-zA-Z-._]+$`,
					MaxLength:    64,
					Message:      "certificate name may contain at most 64 characters and only alphanumeric characters or -._",
				},
			},
			SensitiveFields: []string{
				"properties.certificate",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.isVerified", Value: false},
			},
			RequiredFields: []string{
				"properties.certificate",
			},
		},
	}
}

func init() { azwise.Register(NewIotHubDpsCertificate()) }
