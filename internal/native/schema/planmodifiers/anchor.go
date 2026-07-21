package planmodifiers

// Blank imports anchor the type-specific plan-modifier packages used by
// native-generated resources (RequiresReplace for ForceNew attributes) so that
// `go mod vendor` includes them. The generated files import these directly.
import (
	_ "github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	_ "github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	_ "github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
)
