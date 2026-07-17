package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/typegraph"

// customizeKustoClusterDatabase applies Microsoft.Kusto/clusters/databases schema
// rules the bicep type graph and azwise overlay cannot express:
//   - The database name is an operational-envelope field (not a body property), so
//     the AzureRM DatabaseName rule (validate/name.go) is attached to the envelope
//     name here rather than as an empty-path azwise StringRule (which the overlay
//     skips because name is not in the body graph).
//   - AzureRM uses commonschema.Location() (Required + ForceNew) for a database's
//     location, but bicep marks the ReadWriteDatabase/ReadOnlyFollowing base
//     `location` Optional, so the generator emits it Optional+Computed. Promote the
//     body `location` to Required (azwise RequiredFields is planner metadata, not a
//     schema flag); its ForceNew RequiresReplace already rides the azwise overlay.
func customizeKustoClusterDatabase(def *typegraph.ResourceDefinition) {
	def.Envelope.Name.Validators = []typegraph.DescriptionValidator{
		typegraph.LengthValidator(1, 260),
		typegraph.RegexValidator(
			`^[a-zA-Z0-9\s._-]+$`,
			"Kusto database name must be at most 260 characters of alphanumerics, whitespace, dots, dashes, and underscores",
		),
	}

	typegraph.FindProperty(def, "location").Flags |= typegraph.FlagRequired
}
