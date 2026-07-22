package fluidrelay

import (
	"time"

	"github.com/wuxu92/azwise"
)

// FluidRelayServer provides resource knowledge for
// Microsoft.FluidRelay/fluidRelayServers (Azure Fluid Relay server).
//
// Contributing Terraform resource: azurerm_fluid_relay_server.
//
// Sources:
//   - terraform-provider-azurerm internal/services/fluidrelay/fluid_relay_server_resource.go
//     (schema L75-168, create L178-227, CMK expand/flatten L368-419)
//   - internal/services/fluidrelay/validate/server_name.go (name regex)
//   - go-azure-sdk resource-manager/fluidrelay/2022-05-26/fluidrelayservers:
//     model_fluidrelayserver.go, model_fluidrelayserverproperties.go,
//     model_encryptionproperties.go, constants.go (StorageSKU / CmkIdentityType).
//
// Notes:
//   - name, location & resource_group_name are envelope-owned; name is ForceNew.
//   - storage_sku (ForceNew) maps to properties.storagesku; it is optional+computed so
//     the server picks a value when omitted (DefaultValues with nil value).
//   - customer_managed_key (ForceNew block) maps to properties.encryption; its nested
//     key_vault_key_id (Key Vault nested-item ID) and user_assigned_identity_id (UAI
//     resource ID) are semantic validators ported at the azapin layer, not scalar
//     declarative rules here.
//   - primary_key / secondary_key come from the ListKeys data-plane call, not the ARM
//     body, so they are not modelled as sensitive body fields.
type FluidRelayServer struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*FluidRelayServer)(nil)

func NewFluidRelayServer() *FluidRelayServer {
	return &FluidRelayServer{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.FluidRelay/fluidRelayServers",
			ApiVersions:  []string{"2022-05-26"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 10 * time.Minute,
				Delete: 10 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.storagesku"},
				{PropertyPath: "properties.encryption"},
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.FluidRelayServerName
				{
					Regex:     `^[-0-9a-zA-Z]{1,50}$`,
					MinLength: 1,
					MaxLength: 50,
					Message:   "must contain only alphanumeric characters and hyphens, up to 50 characters long",
				},
				// storage_sku → properties.storagesku
				{
					PropertyPath:  "properties.storagesku",
					AllowedValues: []string{"basic", "standard"},
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// storage_sku is optional+computed — the service decides when omitted.
				{PropertyPath: "properties.storagesku"},
			},
			ComputedFields: []string{
				"properties.frsTenantId",
				"properties.fluidRelayEndpoints",
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewFluidRelayServer()) }
