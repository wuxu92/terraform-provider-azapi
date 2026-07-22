package iothub

import (
	"time"

	"github.com/wuxu92/azwise"
)

// IotHubCertificate provides resource knowledge for
// Microsoft.Devices/IotHubs/certificates.
//
// Contributing Terraform resource: azurerm_iothub_certificate.
//
// Sources:
//   - terraform-provider-azurerm internal/services/iothub/iothub_certificate_resource.go
//     (schema L48-77, Create body L102-107, timeouts L41-46)
//   - terraform-provider-azurerm internal/services/iothub/validate/iot_hub_name.go
//     (IoTHubName L10-19)
//   - go-azure-sdk (kermit) sdk/iothub/2022-04-30-preview/iothub:
//     CertificateDescription / CertificateProperties (Certificate, IsVerified)
//
// Notes:
//   - name/iothub_name/resource_group_name are envelope / parent-reference fields;
//     not emitted as body rules. certificate name reuses the hub-name validator.
//   - certificate_content is Sensitive and maps to properties.certificate.
type IotHubCertificate struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*IotHubCertificate)(nil)

// NewIotHubCertificate returns knowledge for the IotHubs/certificates resource.
func NewIotHubCertificate() *IotHubCertificate {
	return &IotHubCertificate{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Devices/IotHubs/certificates",
			ApiVersions:  []string{"2022-04-30-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[0-9a-zA-Z-]{1,}$`,
					Message:      "certificate name may only contain alphanumeric characters and dashes",
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

func init() { azwise.Register(NewIotHubCertificate()) }
