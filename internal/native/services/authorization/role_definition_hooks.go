package authorization

import (
	"fmt"
	"time"

	"github.com/Azure/terraform-provider-azapi/internal/clients"
	"github.com/Azure/terraform-provider-azapi/internal/native/armjson"
	nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"
)

func init() {
	nativeresource.RegisterHooks(AuthorizationRoleDefinition.Name, &nativeresource.Hooks{
		BeforeCreate: ensureAssignableScopes,
		BeforeUpdate: ensureAssignableScopes,
		AfterUpdate:  settleAfterUpdate,
	})
}

// ensureAssignableScopes guarantees properties.assignableScopes is present and
// non-empty in the ARM request body, defaulting it to the role's own scope (the
// parent_id / ARM ID scope) when the user omitted it. The role definition API
// requires at least one assignable scope; AzureRM supplies [scope] via
// expandRoleDefinitionAssignableScopes, and azapi's mapper otherwise omits an unset
// Optional+Computed list. The scope round-trips without drift: Azure echoes it back
// on read and the attribute's UseStateForUnknown holds it, so a config that omits
// assignable_scopes plans clean after apply.
//
// It is a no-op when properties is absent (schema-required, so ValidateConfig already
// flags its absence) or when the user configured a non-empty assignable_scopes list.
func ensureAssignableScopes(c *nativeresource.CrudCtx) {
	props, ok := c.Body["properties"].(map[string]interface{})
	if !ok {
		return
	}
	scopes, _ := props["assignableScopes"].([]interface{})
	if len(scopes) == 0 {
		props["assignableScopes"] = []interface{}{c.ID.ParentId}
	}
}

// roleDefinitionSettlePoll is how often settleAfterUpdate re-reads the role definition
// while confirming the post-update record has consolidated.
const roleDefinitionSettlePoll = 5 * time.Second

// roleDefinitionSettleContinuous is how many CONSECUTIVE settled reads settleAfterUpdate
// requires before it returns. The consolidation swap does not reach every read replica
// at once, so a single settled GET can be followed by a stale one; requiring a run of
// consecutive settled reads (mirroring AzureRM's ContinuousTargetOccurence) confirms the
// swap has propagated broadly enough that the post-apply refresh reads consistently.
const roleDefinitionSettleContinuous = 12

// settleAfterUpdate blocks until an updated role definition has consolidated in ARM,
// then refreshes c.Response with the settled representation so the flatten — and any
// plan that refreshes right after apply — sees the new values instead of the stale
// pre-update ones.
//
// "Updating" a role definition is not an in-place mutation: ARM writes a NEW record
// whose createdOn and updatedOn both equal the update time, then a few seconds later
// swaps it for the original so createdOn reverts to the role's first-create date and
// only updatedOn carries the update time. Until that swap propagates to the read path,
// a GET can still return the pre-update record (old permissions/description), so the
// immediate GET-after-PUT and an eager refresh report drift. AzureRM waits the same way
// in its Update (roleDefinitionEventualConsistencyUpdate). Create needs no wait: the
// first record is consistent immediately, so only AfterUpdate is registered.
//
// The authoritative update time is properties.updatedOn from the PUT response
// (c.WriteResponse); a GET is settled once its createdOn no longer equals that time
// (the swap happened) and its updatedOn is no older than it (the read reflects our
// write, not a stale record). Because the swap propagates unevenly across read replicas,
// the wait requires roleDefinitionSettleContinuous consecutive settled reads — a single
// stale read resets the run — so it does not return until the consolidated record is
// stably visible. The wait respects the update timeout carried on c.Ctx (azwise default
// 60m); a timeout is surfaced as an error rather than returning stale state.
func settleAfterUpdate(c *nativeresource.CrudCtx) {
	updateRequest, ok := roleDefinitionUpdatedOn(c.WriteResponse)
	if !ok {
		// No authoritative update time to anchor the wait; leave the read-back as-is
		// rather than block indefinitely.
		return
	}

	consecutive := 0
	for {
		if roleDefinitionSettled(c.Response, updateRequest) {
			consecutive++
			if consecutive >= roleDefinitionSettleContinuous {
				return
			}
		} else {
			consecutive = 0
		}

		select {
		case <-c.Ctx.Done():
			c.Diags.AddError(
				"Timed out waiting for role definition to settle",
				fmt.Errorf("waiting for %s to consolidate after update: %w", c.ID.ID(), c.Ctx.Err()).Error(),
			)
			return
		case <-time.After(roleDefinitionSettlePoll):
		}

		respBody, err := c.Client.ResourceClient.Get(c.Ctx, c.ID.AzureResourceId, c.ID.ApiVersion, clients.DefaultRequestOptions())
		if err != nil {
			c.Diags.AddError(
				"Failed to read role definition while settling",
				fmt.Errorf("re-reading %s after update: %w", c.ID.ID(), err).Error(),
			)
			return
		}
		c.Response = armjson.AsMap(respBody)
	}
}

// roleDefinitionSettled reports whether a GET response reflects a consolidated update
// relative to updateRequest (the PUT response's properties.updatedOn). The freshly
// written record has createdOn == updateRequest (an unconsolidated shadow); a stale
// pre-update record has updatedOn before updateRequest. A response is settled only once
// createdOn has reverted (no longer equals updateRequest) and updatedOn is no earlier
// than updateRequest — i.e. the read reflects our write and the swap has completed.
func roleDefinitionSettled(resp map[string]interface{}, updateRequest time.Time) bool {
	createdOn, okC := roleDefinitionTime(resp, "createdOn")
	updatedOn, okU := roleDefinitionUpdatedOn(resp)
	if !okC || !okU {
		return false
	}
	if createdOn.Equal(updateRequest) {
		return false
	}
	return !updatedOn.Before(updateRequest)
}

// roleDefinitionUpdatedOn reads properties.updatedOn from an ARM response map as an
// RFC3339 time.
func roleDefinitionUpdatedOn(resp map[string]interface{}) (time.Time, bool) {
	return roleDefinitionTime(resp, "updatedOn")
}

// roleDefinitionTime reads properties.<field> from an ARM response map and parses it as
// an RFC3339 timestamp. It returns false when the field is absent or unparseable.
func roleDefinitionTime(resp map[string]interface{}, field string) (time.Time, bool) {
	props, ok := resp["properties"].(map[string]interface{})
	if !ok {
		return time.Time{}, false
	}
	raw, ok := props[field].(string)
	if !ok || raw == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}
