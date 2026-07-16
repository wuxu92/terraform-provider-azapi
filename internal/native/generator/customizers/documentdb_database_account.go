package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/typegraph"

// customizeDocumentDBDatabaseAccount applies Microsoft.DocumentDB/databaseAccounts
// schema rules the bicep type graph and azwise overlay cannot express:
//   - The account name is an operational-envelope field (not a body property), so
//     the AzureRM name rule (cosmosdb_account_resource.go:204-211) is attached to
//     the envelope name here rather than as an empty-path azwise StringRule (which
//     the overlay skips because name is not in the body graph).
//
// The value-conditional replacements (analytical-storage disable, backup policy
// Continuous->Periodic) are AzureRM CustomizeDiff rules that depend on the change
// between two values, which a static schema plan modifier cannot express; they live
// in the azwise CheckForceNew override wired through the documentdb ModifyPlan hook.
func customizeDocumentDBDatabaseAccount(def *typegraph.ResourceDefinition) {
	def.Envelope.Name.Validators = []typegraph.DescriptionValidator{
		typegraph.LengthValidator(3, 50),
		typegraph.RegexValidator(
			`^[-a-z0-9]{3,50}$`,
			"Cosmos DB account name must be 3-50 characters of lowercase letters, numbers, and hyphens",
		),
	}
}
