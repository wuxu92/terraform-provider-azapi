// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package mssqlmanagedinstance

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ManagedInstanceAdministrator provides resource knowledge for
// Microsoft.Sql/managedInstances/administrators
// (Terraform azurerm_mssql_managed_instance_active_directory_administrator).
//
// The ARM resource name is the constant "activeDirectory". The body is
// {properties:{administratorType, login, sid, tenantId}} against the go-azure-sdk
// managedinstanceadministrators.ManagedInstanceAdministrator model.
//
// Sources:
//   - terraform-provider-azurerm internal/services/mssqlmanagedinstance/mssql_managed_instance_active_directory_administrator_resource.go:51-83
//     (Arguments: login_username, object_id IsUUID, tenant_id IsUUID)
//   - .../mssql_managed_instance_active_directory_administrator_resource.go:120-127 (Create → properties)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/managedinstanceadministrators:
//     model_managedinstanceadministratorproperties.go:6-11, constants.go
//     (ManagedInstanceAdministratorType = ActiveDirectory only)
//
// Intentionally not encoded (documented, not emitted):
//   - azuread_authentication_only is written to a SEPARATE ARM resource,
//     Microsoft.Sql/managedInstances/azureADOnlyAuthentications (SDK package
//     managedinstanceazureadonlyauthentications, name "Default"), not the
//     administrators body — its rule belongs in that sub-service's file.
//   - properties.administratorType is hardcoded to "ActiveDirectory" by AzureRM and
//     is the only allowed value; it is emitted below as a RequiredField + enum.
type ManagedInstanceAdministrator struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ManagedInstanceAdministrator)(nil)

// NewManagedInstanceAdministrator returns knowledge for Microsoft.Sql/managedInstances/administrators.
func NewManagedInstanceAdministrator() *ManagedInstanceAdministrator {
	return &ManagedInstanceAdministrator{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/managedInstances/administrators",
			ApiVersions:  []string{"2023-08-01-preview"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 180 * time.Minute,
			},
			// managed_instance_id (parent) is envelope-owned RequiresReplace; no body ForceNew.
			ForceNew: []azwise.ForceNewRule{},
			StringRules: []azwise.StringRule{
				// login_username → properties.login (StringIsNotEmpty).
				{PropertyPath: "properties.login", MinLength: 1, Message: "login_username must not be empty"},
				// object_id → properties.sid (validation.IsUUID).
				{
					PropertyPath: "properties.sid",
					Regex:        `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
					Message:      "object_id must be a valid UUID",
				},
				// tenant_id → properties.tenantId (validation.IsUUID).
				{
					PropertyPath: "properties.tenantId",
					Regex:        `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
					Message:      "tenant_id must be a valid UUID",
				},
				// properties.administratorType — AzureRM hardcodes "ActiveDirectory".
				{
					PropertyPath:  "properties.administratorType",
					AllowedValues: []string{"ActiveDirectory"},
					Message:       "must be ActiveDirectory",
				},
			},
			IntRules:        []azwise.IntRule{},
			FloatRules:      []azwise.FloatRule{},
			ArrayRules:      []azwise.ArrayRule{},
			SensitiveFields: []string{},
			ComputedFields:  []string{},
			DefaultValues:   []azwise.DefaultValue{},
			// AzureRM Required:true body fields (login/sid) plus hardcoded administratorType.
			RequiredFields: []string{
				"properties.administratorType",
				"properties.login",
				"properties.sid",
			},
		},
	}
}

func init() { azwise.Register(NewManagedInstanceAdministrator()) }
