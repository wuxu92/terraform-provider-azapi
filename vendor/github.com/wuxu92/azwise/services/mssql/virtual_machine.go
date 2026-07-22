package mssql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualMachine provides resource knowledge for
// Microsoft.SqlVirtualMachine/sqlVirtualMachines.
//
// Contributing Terraform resource: azurerm_mssql_virtual_machine.
//
// Sources:
//   - terraform-provider-azurerm internal/services/mssql/mssql_virtual_machine_resource.go
//     (schema L59-469, Create body L555-582, CustomDiff L484-493)
//   - go-azure-sdk resource-manager/sqlvirtualmachine/2023-10-01/sqlvirtualmachines:
//     model_sqlvirtualmachineproperties.go, model_serverconfigurationsmanagementsettings.go,
//     model_sqlconnectivityupdatesettings.go, model_sqlinstancesettings.go,
//     model_keyvaultcredentialsettings.go, model_storageconfigurationsettings.go, constants.go
//
// Notes:
//   - virtual_machine_id maps to properties.virtualMachineResourceId (the parent VM
//     reference) and is ForceNew.
//   - Nested single-instance blocks are flattened to their ARM object paths. The
//     key_vault_credential sub-fields (key_vault_url/service_principal_name/secret) are
//     ForceNew and Sensitive; sql_instance.collation/instant_file_initialization_enabled/
//     lock_pages_in_memory_enabled are ForceNew.
//   - storage_configuration disk_type / storage_workload_type and auto_patching /
//     auto_backup enums live under those blocks; the ones with a single ARM object path
//     are emitted, array/multi-value collections (auto_backup days_of_week) are skipped.
//   - Removing the auto_backup block is ForceNew via CustomizeDiff; that is a presence
//     transition (block removed) rather than a single value change and is not emitted as a
//     declarative ForceNew path.
type VirtualMachine struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualMachine)(nil)

// NewVirtualMachine returns knowledge for the sqlVirtualMachines resource.
func NewVirtualMachine() *VirtualMachine {
	return &VirtualMachine{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.SqlVirtualMachine/sqlVirtualMachines",
			ApiVersions:  []string{"2023-10-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.virtualMachineResourceId"},
				{PropertyPath: "properties.sqlServerLicenseType"},
				{PropertyPath: "properties.keyVaultCredentialSettings.azureKeyVaultUrl"},
				{PropertyPath: "properties.keyVaultCredentialSettings.servicePrincipalName"},
				{PropertyPath: "properties.keyVaultCredentialSettings.servicePrincipalSecret"},
				{PropertyPath: "properties.serverConfigurationsManagementSettings.sqlInstanceSettings.collation"},
				{PropertyPath: "properties.serverConfigurationsManagementSettings.sqlInstanceSettings.isIfiEnabled"},
				{PropertyPath: "properties.serverConfigurationsManagementSettings.sqlInstanceSettings.isLpimEnabled"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.sqlServerLicenseType",
					AllowedValues: []string{"AHUB", "DR", "PAYG"},
					Message:       "sql_license_type must be AHUB, DR or PAYG",
				},
				{
					PropertyPath:  "properties.serverConfigurationsManagementSettings.sqlConnectivityUpdateSettings.connectivityType",
					AllowedValues: []string{"LOCAL", "PRIVATE", "PUBLIC"},
					Message:       "sql_connectivity_type must be LOCAL, PRIVATE or PUBLIC",
				},
				{
					PropertyPath:  "properties.storageConfigurationSettings.diskConfigurationType",
					AllowedValues: []string{"ADD", "EXTEND", "NEW"},
					Message:       "storage_configuration disk_type must be ADD, EXTEND or NEW",
				},
				{
					PropertyPath:  "properties.storageConfigurationSettings.storageWorkloadType",
					AllowedValues: []string{"DW", "GENERAL", "OLTP"},
					Message:       "storage_configuration storage_workload_type must be DW, GENERAL or OLTP",
				},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.serverConfigurationsManagementSettings.sqlConnectivityUpdateSettings.port", MinValue: azwise.Ptr(int64(1024)), MaxValue: azwise.Ptr(int64(65535))},
				{PropertyPath: "properties.serverConfigurationsManagementSettings.sqlInstanceSettings.maxDop", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(32767))},
				{PropertyPath: "properties.serverConfigurationsManagementSettings.sqlInstanceSettings.maxServerMemoryMB", MinValue: azwise.Ptr(int64(128)), MaxValue: azwise.Ptr(int64(2147483647))},
				{PropertyPath: "properties.serverConfigurationsManagementSettings.sqlInstanceSettings.minServerMemoryMB", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(2147483647))},
			},
			SensitiveFields: []string{
				"properties.keyVaultCredentialSettings.azureKeyVaultUrl",
				"properties.keyVaultCredentialSettings.servicePrincipalName",
				"properties.keyVaultCredentialSettings.servicePrincipalSecret",
				"properties.serverConfigurationsManagementSettings.sqlConnectivityUpdateSettings.sqlAuthUpdatePassword",
				"properties.serverConfigurationsManagementSettings.sqlConnectivityUpdateSettings.sqlAuthUpdateUserName",
				"properties.autoBackupSettings.password",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.serverConfigurationsManagementSettings.sqlConnectivityUpdateSettings.port", Value: 1433},
				{PropertyPath: "properties.serverConfigurationsManagementSettings.sqlConnectivityUpdateSettings.connectivityType", Value: "PRIVATE"},
			},
		},
	}
}

func init() { azwise.Register(NewVirtualMachine()) }
