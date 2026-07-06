package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/typegraph"

// customizeStorageAccountBlobService applies the schema rules the bicep type
// graph cannot express for Microsoft.Storage/storageAccounts/blobServices:
//   - blobServices is a singleton child resource whose ARM name is always
//     "default". The name is not part of the body type graph, so the constraint
//     is attached to the envelope name attribute, rejecting any other value at
//     plan time instead of failing the ARM apply.
func customizeStorageAccountBlobService(def *typegraph.ResourceDefinition) {
	def.Envelope.Name.Validators = []typegraph.DescriptionValidator{typegraph.
		OneOfValidator(
			`blob service name must be "default" (blobServices is a singleton child resource)`,
			"default",
		),
	}

	// Azure normalizes CORS string collections and may return them in a different
	// order than PUT. Model primitive CORS arrays as sets so API ordering does not
	// produce drift. Keep corsRules itself as a list because the max-items rule and
	// object element identity are still list-shaped in the generated schema.
	for _, path := range []string{
		"properties.cors.corsRules.allowedHeaders",
		"properties.cors.corsRules.allowedMethods",
		"properties.cors.corsRules.allowedOrigins",
		"properties.cors.corsRules.exposedHeaders",
	} {
		typegraph.
			FindProperty(def, path).UseSet = true
	}

	// These children are server-controlled/read-only values that Azure leaves null
	// until the sibling policy is enabled, then populates during the same apply:
	//
	//   - lastAccessTimeTrackingPolicy.name -> "AccessTimeTracking"
	//   - lastAccessTimeTrackingPolicy.blobType -> ["blockBlob"]
	//   - lastAccessTimeTrackingPolicy.trackingGranularityInDays -> 1
	//   - restorePolicy.minRestoreTime -> current server time once restore is enabled
	//
	// With the default UseStateForUnknown, the plan pins the stale null and apply fails
	// with "produced an unexpected new value" when Azure supplies the value. Use
	// UseNonNullStateForUnknown so a null prior plans as unknown; once Azure returns a
	// non-null value, later plans reuse that state and stay idempotent.
	for _, path := range []string{
		"properties.lastAccessTimeTrackingPolicy.name",
		"properties.lastAccessTimeTrackingPolicy.blobType",
		"properties.lastAccessTimeTrackingPolicy.trackingGranularityInDays",
		"properties.restorePolicy.minRestoreTime",
	} {
		typegraph.
			FindProperty(def, path).NonNullStateForUnknown = true
	}
}
