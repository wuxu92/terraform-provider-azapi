package springcloud

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SpringCloudApp provides resource knowledge for Microsoft.AppPlatform/Spring/apps.
//
// Mirrors azurerm_spring_cloud_app.
//
// Sources:
//   - terraform-provider-azurerm internal/services/springcloud/spring_cloud_app_resource.go:30-200
//     (Timeouts :49-54 30m/5m/30m/30m; name :57-62 ForceNew; ingress_settings :137-185;
//     is_public :125; https_only :131)
//   - name validate: internal/services/springcloud/validate/spring_cloud_app_name.go:23
//   - go-azure-sdk .../appplatform model_appresourceproperties.go, model_ingresssettings.go,
//     constants.go (BackendProtocol Default/GRPC, SessionAffinity Cookie/None)
type SpringCloudApp struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SpringCloudApp)(nil)

// NewSpringCloudApp returns knowledge for the Spring Cloud app resource.
func NewSpringCloudApp() *SpringCloudApp {
	return &SpringCloudApp{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AppPlatform/Spring/apps",
			ApiVersions:  []string{"2024-01-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// name: begins with a letter, ends alphanumeric, 4-32 lowercase/digits/hyphens.
				{
					PropertyPath: "",
					Regex:        `^([a-z])([a-z\d-]{2,30})([a-z\d])$`,
					Message:      "spring cloud app name must begin with a letter, end with a letter or number, contain only lowercase letters, numbers and hyphens, and be 4-32 characters long",
				},
				// ingress_settings.backend_protocol.
				{
					PropertyPath:  "properties.ingressSettings.backendProtocol",
					AllowedValues: []string{"Default", "GRPC"},
				},
				// ingress_settings.session_affinity.
				{
					PropertyPath:  "properties.ingressSettings.sessionAffinity",
					AllowedValues: []string{"Cookie", "None"},
				},
			},
			IntRules: []azwise.IntRule{
				// ingress_settings.session_cookie_max_age: validation.IntAtLeast(0).
				{PropertyPath: "properties.ingressSettings.sessionCookieMaxAge", MinValue: azwise.Ptr(int64(0))},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.public", Value: false},
				{PropertyPath: "properties.httpsOnly", Value: false},
				{PropertyPath: "properties.ingressSettings.backendProtocol", Value: "Default"},
				{PropertyPath: "properties.ingressSettings.readTimeoutInSeconds", Value: float64(300)},
				{PropertyPath: "properties.ingressSettings.sendTimeoutInSeconds", Value: float64(60)},
				{PropertyPath: "properties.ingressSettings.sessionAffinity", Value: "None"},
			},
		},
	}
}

func init() { azwise.Register(NewSpringCloudApp()) }
