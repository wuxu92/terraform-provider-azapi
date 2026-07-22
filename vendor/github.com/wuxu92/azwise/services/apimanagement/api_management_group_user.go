package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementGroupUser provides resource knowledge for
// Microsoft.ApiManagement/service/groups/users.
//
// Mirrors azurerm_api_management_group_user. This is a pure association resource:
// all four schema attributes (user_id, group_name, resource_group_name,
// api_management_name) are URL/envelope segments and the Create is a bodyless PUT
// (groupuser.Create takes no request body). There are therefore no ARM body
// properties, ForceNew rules, or validation rules to encode beyond timeouts.
//
// Sources (terraform-provider-azurerm internal/services/apimanagement):
//   - api_management_group_user_resource.go (schema L37-45; Create L49-77 — no body)
//   - groupuser/id_groupuser.go (.../groups/{groupId}/users/{userId})
type ApiManagementGroupUser struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementGroupUser)(nil)

// NewApiManagementGroupUser returns knowledge for the groups/users association resource.
func NewApiManagementGroupUser() *ApiManagementGroupUser {
	return &ApiManagementGroupUser{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/groups/users",
			ApiVersions:  []string{"2022-08-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementGroupUser()) }
