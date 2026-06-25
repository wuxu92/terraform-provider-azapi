package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/generator"

// customizeStorageAccountBlobService applies the schema rules the bicep type
// graph cannot express for Microsoft.Storage/storageAccounts/blobServices:
//   - blobServices is a singleton child resource whose ARM name is always
//     "default". The name is not part of the body type graph, so the constraint
//     is attached to the envelope name attribute, rejecting any other value at
//     plan time instead of failing the ARM apply.
func customizeStorageAccountBlobService(def *generator.ResourceDefinition) {
	def.Envelope.Name.Validators = []generator.DescriptionValidator{
		generator.OneOfValidator(
			`blob service name must be "default" (blobServices is a singleton child resource)`,
			"default",
		),
	}

	// lastAccessTimeTrackingPolicy.name is server-controlled and read-only (its
	// only valid value is "AccessTimeTracking"). The server omits it while the
	// policy is disabled and populates it once enabled, so it transitions
	// null -> "AccessTimeTracking". With the default UseStateForUnknown the plan
	// pins the stale null and apply fails with "produced an unexpected new value:
	// ...last_access_time_tracking_policy.name: was null, but now
	// AccessTimeTracking". UseNonNullStateForUnknown plans the null prior as
	// "(known after apply)" so the server may supply the value; once non-null it
	// is reused, keeping the resource idempotent.
	generator.FindProperty(def, "properties.lastAccessTimeTrackingPolicy.name").NonNullStateForUnknown = true
}
