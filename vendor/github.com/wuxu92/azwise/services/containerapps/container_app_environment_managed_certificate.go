package containerapps

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ContainerAppEnvironmentManagedCertificate provides resource knowledge for
// Microsoft.App/managedEnvironments/managedCertificates.
//
// Mirrors azurerm_container_app_environment_managed_certificate.
//
// Sources:
//   - terraform-provider-azurerm internal/services/containerapps/container_app_environment_managed_certificate_resource.go
//     schema (Arguments 47-90, Attributes 92-100) + Create (102-161): name/environment/subject_name/
//     domain_control_validation ForceNew, dcv enum + default HTTP; timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/containerapps/2025-07-01/managedenvironments
//     ManagedCertificateProperties (subjectName/domainControlValidation + read-only validationToken/
//     provisioningState/error), constants.go ManagedCertificateDomainControlValidation enum.
type ContainerAppEnvironmentManagedCertificate struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ContainerAppEnvironmentManagedCertificate)(nil)

func NewContainerAppEnvironmentManagedCertificate() *ContainerAppEnvironmentManagedCertificate {
	return &ContainerAppEnvironmentManagedCertificate{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.App/managedEnvironments/managedCertificates",
			ApiVersions:  []string{"2025-07-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.subjectName"},
				{PropertyPath: "properties.domainControlValidation"},
			},
			RequiredFields: []string{
				"properties.subjectName",
			},
			StringRules: []azwise.StringRule{
				// domain_control_validation -> ManagedCertificateDomainControlValidation enum
				// (full ARM set; AzureRM restricts to CNAME/HTTP).
				{
					PropertyPath:  "properties.domainControlValidation",
					AllowedValues: []string{"CNAME", "HTTP", "TXT"},
					Message:       "domain control validation must be CNAME, HTTP or TXT",
				},
			},
			ComputedFields: []string{
				"properties.validationToken",
				"properties.provisioningState",
				"properties.error",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.domainControlValidation", Value: "HTTP"},
			},
		},
	}
}

func init() { azwise.Register(NewContainerAppEnvironmentManagedCertificate()) }
