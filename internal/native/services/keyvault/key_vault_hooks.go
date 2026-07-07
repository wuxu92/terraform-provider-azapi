package keyvault

import (
	nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"
)

// Key Vault's ARM create/update API requires properties.accessPolicies to be present
// in the request body — it rejects an absent value with "The parameter accessPolicies
// is not specified." AzureRM always sends the slice (empty when the user configures no
// policies), but azapi's mapper omits an unset Optional+Computed list, so a config that
// leaves access_policies out produces a body with no accessPolicies key and the PUT
// fails. This hook injects an empty accessPolicies array whenever the composed body
// lacks one; a user-supplied list is left untouched. The empty list round-trips without
// drift: Azure echoes [] on read and the attribute's UseStateForUnknown holds it.
func init() {
	nativeresource.RegisterHooks(KeyVault.Name, &nativeresource.Hooks{
		BeforeCreate: ensureAccessPolicies,
		BeforeUpdate: ensureAccessPolicies,
	})
}

// ensureAccessPolicies guarantees properties.accessPolicies is present in the ARM
// request body, defaulting it to an empty array when the user omitted it. It is a
// no-op when properties is absent (schema-required, so ValidateConfig already flags
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
