package keyvault

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Azure/terraform-provider-azapi/internal/azure/location"
	"github.com/Azure/terraform-provider-azapi/internal/clients"
	nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"
	"github.com/Azure/terraform-provider-azapi/internal/services/parse"
)

func init() {
	nativeresource.RegisterHooks(KeyVault.Name, &nativeresource.Hooks{
		BeforeCreate: ensureAccessPolicies,
		BeforeUpdate: ensureAccessPolicies,
		AfterDelete:  purgeOnDestroy,
	})
}

// ensureAccessPolicies guarantees properties.accessPolicies is present in the ARM
// request body, defaulting it to an empty array when the user omitted it. Key Vault's
// create/update API rejects an absent value with "The parameter accessPolicies is not
// specified"; AzureRM always sends the slice (empty when no policies), but azapi's
// mapper omits an unset Optional+Computed list. The empty list round-trips without
// drift: Azure echoes [] on read and the attribute's UseStateForUnknown holds it. It is
// a no-op when properties is absent (schema-required, so ValidateConfig already flags
// its absence) or when the user configured access policies.
func ensureAccessPolicies(c *nativeresource.CrudCtx) {
	props, ok := c.Body["properties"].(map[string]interface{})
	if !ok {
		return
	}
	if _, present := props["accessPolicies"]; !present {
		props["accessPolicies"] = []interface{}{}
	}
}

// purgeAction is the outcome of purgePlan: what AfterDelete should do with a vault's
// soft-deleted shadow.
type purgeAction int

const (
	// purgeSkip: purge_on_destroy is unset/false — leave the shadow recoverable.
	purgeSkip purgeAction = iota
	// purgeBlocked: purge requested but purge protection is on, so Azure would reject
	// the purge; warn and leave the shadow for Azure to auto-purge at retention end.
	purgeBlocked
	// purgeMissingLocation: purge requested but state carries no location to locate the
	// location-scoped shadow; warn.
	purgeMissingLocation
	// purgeMissingSubscription: purge requested but the ARM ID has no subscription; this
	// is a malformed-ID error, not an expected skip.
	purgeMissingSubscription
	// purgeDo: purge the shadow at DeletedVaultID.
	purgeDo
)

// purgePlan decides, purely from the just-deleted vault's prior state and ID, whether
// and how AfterDelete should purge the soft-deleted shadow. It performs no I/O so the
// full decision matrix is unit-testable; purgeOnDestroy interprets the result and makes
// the single live purge call. DeletedVaultID is set only for purgeDo.
func purgePlan(state types.Object, id parse.ResourceId) (action purgeAction, deletedVaultID string) {
	if !nativeresource.AttrBool(state, "purge_on_destroy") {
		return purgeSkip, ""
	}
	if nativeresource.AttrBool(nativeresource.AttrObject(state, "properties"), "enable_purge_protection") {
		return purgeBlocked, ""
	}
	loc := location.Normalize(nativeresource.AttrString(state, "location"))
	if loc == "" {
		return purgeMissingLocation, ""
	}
	subscriptionID := id.SubscriptionId()
	if subscriptionID == "" {
		return purgeMissingSubscription, ""
	}
	return purgeDo, fmt.Sprintf(
		"/subscriptions/%s/providers/Microsoft.KeyVault/locations/%s/deletedVaults/%s",
		subscriptionID, loc, id.Name,
	)
}

// purgeOnDestroy permanently purges a vault's soft-deleted shadow after the ARM DELETE
// when the practitioner set purge_on_destroy = true. Key Vault soft-delete is always on
// for modern vaults, so a plain DELETE only moves the vault into a recoverable deleted
// state that keeps reserving its name until Azure's retention window expires; purging
// frees the name immediately. The purge targets the location-scoped deletedVaults
// resource (which only exists after the DELETE), so this runs in AfterDelete, not
// BeforeDelete.
//
// The decision (purgePlan) is pure; this wrapper turns it into diagnostics and the one
// live purge call. A purge failure is surfaced as an error; the vault is already
// soft-deleted, so a re-run's DELETE 404s and reaches the purge again.
func purgeOnDestroy(c *nativeresource.CrudCtx) {
	action, deletedVaultID := purgePlan(c.State, c.ID)
	switch action {
	case purgeSkip:
		return
	case purgeBlocked:
		c.Diags.AddWarning(
			"Key Vault not purged",
			fmt.Sprintf("purge_on_destroy is set but %s has purge protection enabled; Azure blocks purging a "+
				"protected vault, so its soft-deleted shadow is retained until the retention window expires.", c.ID.ID()),
		)
		return
	case purgeMissingLocation:
		c.Diags.AddWarning(
			"Key Vault not purged",
			fmt.Sprintf("purge_on_destroy is set but %s has no location in state; cannot locate its soft-deleted "+
				"shadow to purge.", c.ID.ID()),
		)
		return
	case purgeMissingSubscription:
		c.Diags.AddError(
			"Failed to purge Key Vault",
			fmt.Sprintf("could not determine the subscription of %q to locate its soft-deleted shadow", c.ID.AzureResourceId),
		)
		return
	}

	if _, err := c.Client.ResourceClient.Action(
		c.Ctx, deletedVaultID, "purge", c.ID.ApiVersion, http.MethodPost, nil, clients.DefaultRequestOptions(),
	); err != nil {
		c.Diags.AddError(
			"Failed to purge Key Vault",
			fmt.Errorf("purging soft-deleted %s: %w", c.ID.ID(), err).Error(),
		)
	}
}
