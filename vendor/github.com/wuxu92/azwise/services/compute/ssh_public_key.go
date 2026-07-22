package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SshPublicKey provides resource knowledge for Microsoft.Compute/sshPublicKeys.
//
// Contributing Terraform resource: azurerm_ssh_public_key.
//
// Sources:
//   - terraform-provider-azurerm internal/services/compute/ssh_public_key_resource.go
//     (schema L26-70, Create body L93-99)
//   - terraform-provider-azurerm internal/services/compute/validate/ssh_key.go (public_key)
//   - go-azure-sdk resource-manager/compute/2024-03-01/sshpublickeys:
//     model_sshpublickeyresourceproperties.go.
//
// Notes:
//   - name/location/resource_group_name are envelope-owned; name is ForceNew but is not an
//     ARM-body property, so only its naming regex is emitted (empty PropertyPath).
//   - public_key (Required, updatable) maps to properties.publicKey; validate.SSHKey is a
//     semantic OpenSSH-format check that cannot be expressed as a StringRule regex —
//     left as a note for an azapin service validator.
type SshPublicKey struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SshPublicKey)(nil)

// NewSshPublicKey returns knowledge for the sshPublicKeys resource.
func NewSshPublicKey() *SshPublicKey {
	return &SshPublicKey{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/sshPublicKeys",
			ApiVersions:  []string{"2024-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 45 * time.Minute,
				Read:   5 * time.Minute,
				Update: 45 * time.Minute,
				Delete: 45 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// Resource name validation (empty PropertyPath validates the name).
				{Regex: `^[-a-zA-Z0-9(_).]{1,128}$`, MinLength: 1, MaxLength: 128, Message: "must be 1-128 characters and may contain letters, numbers, underscores, dots, parentheses and hyphens"},
			},
			RequiredFields: []string{
				"properties.publicKey",
			},
		},
	}
}

func init() { azwise.Register(NewSshPublicKey()) }
