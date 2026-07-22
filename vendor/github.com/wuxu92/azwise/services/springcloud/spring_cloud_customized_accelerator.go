package springcloud

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SpringCloudCustomizedAccelerator provides resource knowledge for
// Microsoft.AppPlatform/Spring/applicationAccelerators/customizedAccelerators.
//
// Mirrors azurerm_spring_cloud_customized_accelerator.
//
// Sources:
//   - internal/services/springcloud/spring_cloud_customized_accelerator_resource.go:93-220
//     (Timeouts Create :257 30m / Update :304 30m / Read :367 5m / Delete :421 30m;
//     name :94-99 ForceNew; git_repository.basic_auth :115-137 ForceNew;
//     git_repository.ssh_auth :139-168 ForceNew; interval_in_seconds :197-201 IntAtLeast(10))
//   - go-azure-sdk .../appplatform model_customizedacceleratorproperties.go (GitRepository json
//     "gitRepository"), model_acceleratorgitrepository.go (AuthSetting "authSetting",
//     IntervalInSeconds "intervalInSeconds")
type SpringCloudCustomizedAccelerator struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SpringCloudCustomizedAccelerator)(nil)

// NewSpringCloudCustomizedAccelerator returns knowledge for the customized accelerator resource.
func NewSpringCloudCustomizedAccelerator() *SpringCloudCustomizedAccelerator {
	return &SpringCloudCustomizedAccelerator{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AppPlatform/Spring/applicationAccelerators/customizedAccelerators",
			ApiVersions:  []string{"2024-01-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// git_repository auth settings are chosen at create time and are replace-only.
			// The auth setting is a discriminated union (properties.gitRepository.authSetting);
			// both basic_auth and ssh_auth map onto it, so it is listed once.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.gitRepository.authSetting"},
			},
			IntRules: []azwise.IntRule{
				// git_repository.interval_in_seconds: validation.IntAtLeast(10).
				{PropertyPath: "properties.gitRepository.intervalInSeconds", MinValue: azwise.Ptr(int64(10))},
			},
			// git_repository.basic_auth.password / ssh_auth.private_key / host_key are sensitive
			// but live under the discriminated authSetting union, so they are documented rather
			// than emitted as declarative SensitiveFields.
		},
	}
}

func init() { azwise.Register(NewSpringCloudCustomizedAccelerator()) }
