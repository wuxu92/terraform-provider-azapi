package authorization

import (
	"fmt"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/native/services/config"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
)

// TestRoleAssignmentCfgBasic pins the HCL contract for the native role-assignment
// acceptance scenario: a subscription-scope azapi_role_assignment grants the built-in
// Reader role to data.azapi_client_config.current.object_id using the native snake_case
// body fields, not the generic azapi_resource body envelope.
func TestRoleAssignmentCfgBasic(t *testing.T) {
	cfg := NewRoleAssignmentCfg(config.ClientConfig, "reader")
	cfg.name = "00000000-0000-0000-0000-000000000000"

	got := RoleAssignmentCfg_Basic(cfg).Config()
	want := `
resource "azapi_role_assignment" "reader" {
  name     = "00000000-0000-0000-0000-000000000000"
  scope_id = "/subscriptions/${data.azapi_client_config.current.subscription_id}"
  properties = {
    principal_id       = data.azapi_client_config.current.object_id
    role_definition_id = "/subscriptions/${data.azapi_client_config.current.subscription_id}/providers/Microsoft.Authorization/roleDefinitions/acdd72a7-3385-48ef-bd42-f606fba81ae7"
  }
}
`

	if got != want {
		t.Errorf("role assignment basic config mismatch\n got: %q\nwant: %q", got, want)
	}
}

func TestKeyVaultDataPlaneRoleAssignmentCfgs(t *testing.T) {
	rg := resources.NewResourceGroupCfg("rg")
	const assignmentID = "11111111-1111-1111-1111-111111111111"

	cases := []struct {
		name      string
		label     string
		roleDefID string
		render    func() string
	}{
		{
			name:      "administrator",
			label:     "key_vault_administrator",
			roleDefID: keyVaultAdministratorRoleDefinitionID,
			render: func() string {
				cfg := NewKeyVaultAdministratorRoleAssignmentCfg(rg, config.ClientConfig)
				cfg.name = assignmentID
				return cfg.Config()
			},
		},
		{
			name:      "certificates officer",
			label:     "key_vault_certificates_officer",
			roleDefID: keyVaultCertificatesOfficerRoleDefinitionID,
			render: func() string {
				cfg := NewKeyVaultCertificatesOfficerRoleAssignmentCfg(rg, config.ClientConfig)
				cfg.name = assignmentID
				return cfg.Config()
			},
		},
		{
			name:      "crypto officer",
			label:     "key_vault_crypto_officer",
			roleDefID: keyVaultCryptoOfficerRoleDefinitionID,
			render: func() string {
				cfg := NewKeyVaultCryptoOfficerRoleAssignmentCfg(rg, config.ClientConfig)
				cfg.name = assignmentID
				return cfg.Config()
			},
		},
		{
			name:      "secrets officer",
			label:     "key_vault_secrets_officer",
			roleDefID: keyVaultSecretsOfficerRoleDefinitionID,
			render: func() string {
				cfg := NewKeyVaultSecretsOfficerRoleAssignmentCfg(rg, config.ClientConfig)
				cfg.name = assignmentID
				return cfg.Config()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.render()
			want := fmt.Sprintf(`
resource "azapi_role_assignment" %q {
  name     = %q
  scope_id = azapi_resource_group.rg.id
  properties = {
    principal_id       = data.azapi_client_config.current.object_id
    role_definition_id = "/subscriptions/${data.azapi_client_config.current.subscription_id}/providers/Microsoft.Authorization/roleDefinitions/%s"
  }
}
`, tc.label, assignmentID, tc.roleDefID)
			if got != want {
				t.Errorf("%s config mismatch\n got: %q\nwant: %q", tc.name, got, want)
			}
		})
	}
}
