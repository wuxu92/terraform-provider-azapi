package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementUser provides resource knowledge for
// Microsoft.ApiManagement/service/users.
//
// Mirrors azurerm_api_management_user. firstName/lastName/email are required.
// `confirmation` is ForceNew. `password` is sensitive. `state` and
// `confirmation` are enums.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_user_resource.go
//     schema (lines 41-97); Create body (lines 131-151); Timeouts 45m/5m/45m/45m.
//   - Microsoft.ApiManagement/service/users@2022-08-01 user.UserCreateParameterProperties:
//     firstName/lastName/email (required), confirmation (enum invite|signup),
//     note, password, state (enum active|blocked|deleted|pending).
type ApiManagementUser struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementUser)(nil)

// NewApiManagementUser returns knowledge for the API Management user resource.
func NewApiManagementUser() *ApiManagementUser {
	return &ApiManagementUser{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/users",
			ApiVersions:  []string{"2022-08-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.confirmation"},
			},
			RequiredFields: []string{
				"properties.firstName",
				"properties.lastName",
				"properties.email",
			},
			SensitiveFields: []string{
				"properties.password",
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.confirmation",
					AllowedValues: []string{"invite", "signup"},
					Message:       "confirmation must be one of invite, signup",
				},
				{
					PropertyPath:  "properties.state",
					AllowedValues: []string{"active", "blocked", "deleted", "pending"},
					Message:       "state must be one of active, blocked, deleted, pending",
				},
				{PropertyPath: "properties.firstName", MinLength: 1, Message: "first_name must not be empty"},
				{PropertyPath: "properties.lastName", MinLength: 1, Message: "last_name must not be empty"},
				{PropertyPath: "properties.email", MinLength: 1, Message: "email must not be empty"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 45 * time.Minute,
				Read:   5 * time.Minute,
				Update: 45 * time.Minute,
				Delete: 45 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementUser()) }
