package springcloud

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SpringCloudCustomDomain provides resource knowledge for
// Microsoft.AppPlatform/Spring/apps/domains.
//
// Mirrors azurerm_spring_cloud_custom_domain.
//
// Sources:
//   - internal/services/springcloud/spring_cloud_custom_domain_resource.go:25-80
//     (Timeouts :44-49 30m/5m/30m/30m; name :52-57 ForceNew; certificate_name :66-71;
//     thumbprint :73-79 ForceNew + RequiredWith certificate_name)
//   - name validate: internal/services/springcloud/validate/spring_cloud_custom_domain_name.go:18
//   - go-azure-sdk .../appplatform model_customdomainproperties.go (CertName json "certName",
//     Thumbprint "thumbprint")
type SpringCloudCustomDomain struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SpringCloudCustomDomain)(nil)

// NewSpringCloudCustomDomain returns knowledge for the Spring Cloud custom domain resource.
func NewSpringCloudCustomDomain() *SpringCloudCustomDomain {
	return &SpringCloudCustomDomain{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AppPlatform/Spring/apps/domains",
			ApiVersions:  []string{"2024-01-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.thumbprint"},
			},
			StringRules: []azwise.StringRule{
				// name: a valid domain name.
				{
					PropertyPath: "",
					Regex:        `^([a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}$`,
					Message:      "spring cloud custom domain name must be a valid domain name",
				},
			},
			// thumbprint is only valid together with certificate_name (certName).
			RequiredWith: []azwise.RelationalRule{
				{Paths: []string{"properties.thumbprint", "properties.certName"}},
			},
		},
	}
}

func init() { azwise.Register(NewSpringCloudCustomDomain()) }
