package mssql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualMachineGroup provides resource knowledge for
// Microsoft.SqlVirtualMachine/sqlVirtualMachineGroups.
//
// Contributing Terraform resource: azurerm_mssql_virtual_machine_group.
//
// Sources:
//   - terraform-provider-azurerm internal/services/mssql/mssql_virtual_machine_group_resource.go
//     (schema L65-164, Create body L196-205)
//   - go-azure-sdk resource-manager/sqlvirtualmachine/2023-10-01/sqlvirtualmachinegroups:
//     model_sqlvirtualmachinegroupproperties.go, model_wsfcdomainprofile.go, constants.go
//     (SqlVMGroupImageSku Developer/Enterprise, ClusterSubnetType MultiSubnet/SingleSubnet)
//
// Notes:
//   - name/resource_group_name/location are envelope-owned; not emitted as body rules.
//   - sql_image_offer is ForceNew (validate.SqlImageOfferName, semantic; no declarative
//     value rule). sql_image_sku is a mutable enum.
//   - wsfc_domain_profile is a single-instance block flattened to
//     properties.wsfcDomainProfile.*; its cluster_subnet_type enum and ForceNew apply to
//     the flattened object path. fqdn/organizational_unit_path/account names are ForceNew
//     but are free-form strings (StringIsNotEmpty), so only the enum + subnet-type
//     ForceNew are emitted declaratively.
type VirtualMachineGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualMachineGroup)(nil)

// NewVirtualMachineGroup returns knowledge for the sqlVirtualMachineGroups resource.
func NewVirtualMachineGroup() *VirtualMachineGroup {
	return &VirtualMachineGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.SqlVirtualMachine/sqlVirtualMachineGroups",
			ApiVersions:  []string{"2023-10-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.sqlImageOffer"},
				{PropertyPath: "properties.wsfcDomainProfile.clusterSubnetType"},
				{PropertyPath: "properties.wsfcDomainProfile.domainFqdn"},
				{PropertyPath: "properties.wsfcDomainProfile.ouPath"},
				{PropertyPath: "properties.wsfcDomainProfile.clusterBootstrapAccount"},
				{PropertyPath: "properties.wsfcDomainProfile.clusterOperatorAccount"},
				{PropertyPath: "properties.wsfcDomainProfile.sqlServiceAccount"},
				{PropertyPath: "properties.wsfcDomainProfile.storageAccountUrl"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					MinLength:    1,
					MaxLength:    15,
					Message:      "virtual machine group name must be 1-15 characters",
				},
				{
					PropertyPath:  "properties.sqlImageSku",
					AllowedValues: []string{"Developer", "Enterprise"},
					Message:       "sql_image_sku must be Developer or Enterprise",
				},
				{
					PropertyPath:  "properties.wsfcDomainProfile.clusterSubnetType",
					AllowedValues: []string{"MultiSubnet", "SingleSubnet"},
					Message:       "cluster_subnet_type must be MultiSubnet or SingleSubnet",
				},
			},
			SensitiveFields: []string{
				"properties.wsfcDomainProfile.storageAccountPrimaryKey",
			},
			RequiredFields: []string{
				"properties.sqlImageOffer",
				"properties.sqlImageSku",
				"properties.wsfcDomainProfile.clusterSubnetType",
				"properties.wsfcDomainProfile.domainFqdn",
			},
		},
	}
}

func init() { azwise.Register(NewVirtualMachineGroup()) }
