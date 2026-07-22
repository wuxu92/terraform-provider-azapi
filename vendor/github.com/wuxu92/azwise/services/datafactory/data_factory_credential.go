package datafactory

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DataFactoryCredential provides resource knowledge for
// Microsoft.DataFactory/factories/credentials.
//
// This ARM type is split by AzureRM into two typed Terraform resources,
// discriminated by the ARM body's properties.type field:
//   - azurerm_data_factory_credential_service_principal      -> properties.type = "ServicePrincipal"
//   - azurerm_data_factory_credential_user_managed_identity -> properties.type = "ManagedIdentity"
//
// Only knowledge universal across both contributors is unioned here. Both are
// implemented as typed (sdk.Resource) resources with a flat 5-minute timeout on
// every CRUD op. The kind-specific payload lives inside the discriminated
// properties.typeProperties union, whose paths do not resolve through the SDK's
// interface-typed Properties field.
//
// Sources:
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_credential_service_principal_resource.go:47-195
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_credential_user_managed_identity_resource.go:41-95
//   - terraform-provider-azurerm vendor/.../datafactory/2018-06-01/credentials/model_credential.go:12-26
//     (Credential interface; ServicePrincipalCredential / ManagedIdentityCredential)
//
// Intentionally skipped here:
//   - name: both use validation.StringIsNotEmpty — a non-empty check, not
//     expressible as a declarative StringRule (no regex/enum/length).
//   - data_factory_id: parent reference (AzAPI ID segment), not a body property.
//   - Kind-specific fields (non-universal, would corrupt the sibling kind if unioned):
//       * ServicePrincipal: tenant_id / service_principal_id validation.IsUUID
//         (semantic — route to an azapin customizer validator, not a StringRule),
//         service_principal_key.{linked_service_name,secret_name,secret_version}
//         inside properties.typeProperties.
//       * ManagedIdentity: identity_id (ForceNew, commonids user-assigned identity
//         resource ID) -> properties.typeProperties.resourceId.
//     These live under the interface Properties field and/or are kind-only, so they
//     are documented rather than emitted.
type DataFactoryCredential struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DataFactoryCredential)(nil)

// NewDataFactoryCredential returns merged knowledge for the
// Microsoft.DataFactory/factories/credentials resource.
func NewDataFactoryCredential() *DataFactoryCredential {
	return &DataFactoryCredential{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DataFactory/factories/credentials",
			ApiVersions:  []string{"2018-06-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 5 * time.Minute,
				Read:   5 * time.Minute,
				Update: 5 * time.Minute,
				Delete: 5 * time.Minute,
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDataFactoryCredential()) }
