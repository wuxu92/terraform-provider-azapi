package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualMachineRunCommand provides resource knowledge for
// Microsoft.Compute/virtualMachines/runCommands.
//
// Contributing TF resource:
//   - azurerm_virtual_machine_run_command (internal/services/compute/virtual_machine_run_command_resource.go)
//
// Sources:
//   - virtual_machine_run_command_resource.go:88-350 (schema), :352-580 (timeouts)
//   - internal/services/compute/validate/virtual_machine_run_command_name.go (name validation)
//   - go-azure-sdk compute/2023-03-01/virtualmachineruncommands/model_virtualmachineruncommandproperties.go
//
// Not emitted:
//   - source.command_id / script / script_uri ExactlyOneOf and
//     source.script_uri_managed_identity RequiredWith source.script_uri live under an
//     array element (properties.source is a single object but the constraint is nested
//     inside the source block via TF array indexing) — array-element/nested paths that
//     the generator cannot lower, so they are skipped.
type VirtualMachineRunCommand struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualMachineRunCommand)(nil)

func NewVirtualMachineRunCommand() *VirtualMachineRunCommand {
	return &VirtualMachineRunCommand{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/virtualMachines/runCommands",
			ApiVersions:  []string{"2023-03-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── VirtualMachineRunCommandName: 1-80 chars, alphanumerics/dots/dashes/underscores.
				{
					Regex:     `^[a-zA-Z0-9._-]+$`,
					MinLength: 1,
					MaxLength: 80,
					Message:   "must be 1-80 characters and may only contain alphanumeric characters, dots, dashes and underscores",
				},
			},
			SensitiveFields: []string{
				"properties.runAsPassword",       // run_as_password (Sensitive)
				"properties.protectedParameters", // protected_parameter (Sensitive)
			},
			// instance_view is Computed — populated by Azure, absent from the create/update model.
			ComputedFields: []string{
				"properties.instanceView",
			},
			// error_blob_managed_identity RequiredWith error_blob_uri,
			// output_blob_managed_identity RequiredWith output_blob_uri.
			RequiredWith: []azwise.RelationalRule{
				{Paths: []string{"properties.errorBlobManagedIdentity", "properties.errorBlobUri"}},
				{Paths: []string{"properties.outputBlobManagedIdentity", "properties.outputBlobUri"}},
			},
		},
	}
}

func init() { azwise.Register(NewVirtualMachineRunCommand()) }
