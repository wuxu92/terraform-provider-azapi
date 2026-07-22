package mssql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualMachineAvailabilityGroupListener provides resource knowledge for
// Microsoft.SqlVirtualMachine/sqlVirtualMachineGroups/availabilityGroupListeners.
//
// Contributing Terraform resource:
// azurerm_mssql_virtual_machine_availability_group_listener.
//
// Sources:
//   - terraform-provider-azurerm internal/services/mssql/mssql_virtual_machine_availability_group_listener_resource.go
//     (schema L79-233, Create body L275-296)
//   - go-azure-sdk resource-manager/sqlvirtualmachine/2023-10-01/availabilitygrouplisteners:
//     model_availabilitygrouplistenerproperties.go, constants.go
//     (Role, Commit, Failover, ReadableSecondary)
//
// Notes:
//   - name/sql_virtual_machine_group_id are envelope/parent references; not emitted as
//     body rules. The entire resource is effectively immutable (all fields ForceNew) and
//     has no Update method.
//   - load_balancer_configuration and multi_subnet_ip_configuration are ExactlyOneOf; both
//     map to array body properties (loadBalancerConfigurations / multiSubnetIpConfigurations)
//     — emitted as an ExactlyOneOf relational rule on the array-presence paths.
//   - replica[*] and the load-balancer/multi-subnet element fields (role/commit/
//     failover_mode/readable_secondary enums, IP addresses) are array-element paths that
//     azwise declarative rules do not support and are intentionally not emitted.
type VirtualMachineAvailabilityGroupListener struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualMachineAvailabilityGroupListener)(nil)

// NewVirtualMachineAvailabilityGroupListener returns knowledge for the
// availabilityGroupListeners resource.
func NewVirtualMachineAvailabilityGroupListener() *VirtualMachineAvailabilityGroupListener {
	return &VirtualMachineAvailabilityGroupListener{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.SqlVirtualMachine/sqlVirtualMachineGroups/availabilityGroupListeners",
			ApiVersions:  []string{"2023-10-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.availabilityGroupName"},
				{PropertyPath: "properties.port"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					MinLength:    1,
					MaxLength:    15,
					Message:      "availability group listener name must be 1-15 characters",
				},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.port", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(65535))},
			},
			ExactlyOneOf: []azwise.RelationalRule{
				{Paths: []string{"properties.loadBalancerConfigurations", "properties.multiSubnetIpConfigurations"}},
			},
		},
	}
}

func init() { azwise.Register(NewVirtualMachineAvailabilityGroupListener()) }
