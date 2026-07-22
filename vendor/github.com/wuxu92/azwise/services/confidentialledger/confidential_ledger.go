package confidentialledger

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ConfidentialLedger provides resource knowledge for
// Microsoft.ConfidentialLedger/ledgers.
//
// Mirrors azurerm_confidential_ledger. name/location/resource_group_name are
// envelope fields.
//
// Notes:
//   - ledger_type is Required + ForceNew -> properties.ledgerType (enum).
//   - azuread_based_service_principal is Required (>=1) ->
//     properties.aadBasedSecurityPrincipals.
//   - The per-principal fields (ledger_role_name enum; principal_id / tenant_id
//     UUID; pem_public_key non-empty) live on array elements
//     (properties.aadBasedSecurityPrincipals[*].* /
//     properties.certBasedSecurityPrincipals[*].*); array-element paths are not
//     expressible as declarative rules, so they are documented here but not
//     emitted.
//   - identity_service_endpoint / ledger_endpoint plus provisioningState,
//     ledgerName and ledgerInternalNamespace are server-computed read-only.
//
// Sources:
//   - terraform-provider-azurerm internal/services/confidentialledger/confidential_ledger_resource.go
//     (schema L45-133: name ForceNew; ledger_type Required+ForceNew StringInSlice;
//     aad principals Required MinItems 1; cert principals Optional; create L136-176;
//     timeouts 30m/5m/30m/30m)
//   - internal/services/confidentialledger/validate/confidential_ledger_name.go
//     (max 32 chars, regex ^[^\-][A-Za-z0-9\-]{1,33}[^\-]$, no leading/trailing dash)
//   - go-azure-sdk resource-manager/confidentialledger/2022-05-13/confidentialledger
//     LedgerProperties{aadBasedSecurityPrincipals,certBasedSecurityPrincipals,
//     ledgerType,identityServiceUri,ledgerUri,ledgerName,ledgerInternalNamespace,
//     provisioningState}; LedgerType = Private | Public | Unknown
type ConfidentialLedger struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ConfidentialLedger)(nil)

// NewConfidentialLedger returns knowledge for the ledgers resource.
func NewConfidentialLedger() *ConfidentialLedger {
	return &ConfidentialLedger{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ConfidentialLedger/ledgers",
			ApiVersions:  []string{"2022-05-13"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.ledgerType"},
			},
			StringRules: []azwise.StringRule{
				// name (ConfidentialLedgerName) validates the ARM resource name.
				{Regex: `^[^\-][A-Za-z0-9\-]{1,33}[^\-]$`, MaxLength: 32, Message: "name may only contain alphanumeric characters and dashes, may not start or end with a dash, and may not exceed 32 characters"},
				// ledger_type -> properties.ledgerType (full ARM LedgerType set).
				{PropertyPath: "properties.ledgerType", AllowedValues: []string{"Private", "Public", "Unknown"}},
			},
			RequiredFields: []string{
				"properties.ledgerType",
				"properties.aadBasedSecurityPrincipals",
			},
			ComputedFields: []string{
				"properties.identityServiceUri",
				"properties.ledgerUri",
				"properties.ledgerName",
				"properties.ledgerInternalNamespace",
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewConfidentialLedger()) }
