package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApplicationSecurityGroup provides resource knowledge for
// Microsoft.Network/applicationSecurityGroups.
//
// Mirrors azurerm_application_security_group. The resource carries only the
// operational envelope (name, resource group) plus location and tags — there is
// no settable body beyond an empty properties object, so only location ForceNew
// and timeouts are encoded.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/application_security_group_resource.go
//     (resourceApplicationSecurityGroup schema + timeouts, lines 28-61)
//   - go-azure-sdk resource-manager/network/2025-01-01/applicationsecuritygroups:
//     id_applicationsecuritygroup.go (ARM type segment "applicationSecurityGroups")
type ApplicationSecurityGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApplicationSecurityGroup)(nil)

// NewApplicationSecurityGroup returns knowledge for the applicationSecurityGroups resource.
func NewApplicationSecurityGroup() *ApplicationSecurityGroup {
	return &ApplicationSecurityGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/applicationSecurityGroups",
			ApiVersions:  []string{"2025-01-01"},
			// location (commonschema.Location) replaces the resource on change;
			// name and resource_group are envelope-owned and already RequiresReplace.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewApplicationSecurityGroup()) }
