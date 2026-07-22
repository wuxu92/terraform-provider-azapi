package codesigning

import (
	"time"

	"github.com/wuxu92/azwise"
)

// TrustedSigningAccount provides resource knowledge for
// Microsoft.CodeSigning/codeSigningAccounts.
//
// Mirrors azurerm_trusted_signing_account. name/location/resource_group_name are
// envelope fields.
//
// Notes:
//   - sku_name is Required -> properties.sku.name (enum). It is updatable (patched
//     via CodeSigningAccountPatchProperties), so it is not ForceNew.
//   - account_uri (properties.accountUri) and provisioningState are server-computed
//     read-only.
//
// Sources:
//   - terraform-provider-azurerm internal/services/codesigning/trusted_signing_account_resource.go
//     (arguments L35-65: name ForceNew len 3-24 + regex; sku_name Required
//     StringInSlice PossibleValuesForSkuName; attributes L67-74 account_uri computed;
//     create L84-127 timeout 30m; read 5m; update 10m; delete 10m)
//   - go-azure-sdk resource-manager/codesigning/2024-09-30-preview/codesigningaccounts
//     CodeSigningAccountProperties{sku,accountUri,provisioningState};
//     AccountSku.Name; SkuName = Basic | Premium
type TrustedSigningAccount struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*TrustedSigningAccount)(nil)

// NewTrustedSigningAccount returns knowledge for the codeSigningAccounts resource.
func NewTrustedSigningAccount() *TrustedSigningAccount {
	return &TrustedSigningAccount{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.CodeSigning/codeSigningAccounts",
			ApiVersions:  []string{"2024-09-30-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 10 * time.Minute,
				Delete: 10 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// name validates the ARM resource name.
				{Regex: `^[A-Za-z][A-Za-z0-9]*(?:-[A-Za-z0-9]+)*$`, MinLength: 3, MaxLength: 24, Message: "name must be 3-24 alphanumeric characters, begin with a letter, end with a letter or digit, and not contain consecutive hyphens"},
				// sku_name -> properties.sku.name (full ARM SkuName set).
				{PropertyPath: "properties.sku.name", AllowedValues: []string{"Basic", "Premium"}},
			},
			RequiredFields: []string{
				"properties.sku.name",
			},
			ComputedFields: []string{
				"properties.accountUri",
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewTrustedSigningAccount()) }
