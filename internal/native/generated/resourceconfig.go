package generated

// Terraform resource type names of the native static resources — the values of
// Descriptor.Name. Like armtypes for ARM types, these are the single source of
// truth for the "azapi_*" type strings: reference the constant (e.g. from an
// acceptance-test config builder's NewXxx constructor) instead of repeating the
// literal.
const (
	TypeResourceGroup             = "azapi_resource_group"
	TypeStorageAccount            = "azapi_storage_account"
	TypeStorageAccountBlobService = "azapi_storage_account_blob_service"
)

// ResourceConfigBase carries the Terraform type and state label shared by every
// acceptance-test config builder. A builder embeds it (constructing it via
// NewResourceConfigBase in its NewXxx constructor) to get ResourceType and
// ResourceLabel for free and thereby satisfy nativeacc.ResourceConfig, so a
// scope's ResourceFor can vend the matching Resource handle. The type is set
// once, from a Type* constant, in the constructor — callers never restate it.
//
// It lives here (not in nativeacc) because the config builders are in the
// generated service packages, which nativeacc transitively imports: embedding a
// nativeacc type would form an import cycle. The interface is structural, so
// satisfying nativeacc.ResourceConfig needs no import of nativeacc.
type ResourceConfigBase struct {
	tfType string
	label  string
}

// NewResourceConfigBase builds the embedded base for a config builder's NewXxx
// constructor: tfType is a Type* constant, label is the Terraform state label.
func NewResourceConfigBase(tfType, label string) ResourceConfigBase {
	return ResourceConfigBase{tfType: tfType, label: label}
}

// ResourceType returns the Terraform type (implements nativeacc.ResourceConfig).
func (b ResourceConfigBase) ResourceType() string { return b.tfType }

// ResourceLabel returns the Terraform state label (implements nativeacc.ResourceConfig).
func (b ResourceConfigBase) ResourceLabel() string { return b.label }
