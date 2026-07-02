package services

// ResourceConfigBase carries the Terraform type and state label shared by every
// acceptance-test config builder. A builder embeds it (constructing it via
// NewResourceConfigBase in its NewXxxCfg constructor) to get ResourceType and
// ResourceLabel for free and thereby satisfy nativeacc.ResourceConfig, so a
// scope's ResourceFor can vend the matching Resource handle. The type is set
// once, from the resource's generated Descriptor.Name — callers never restate it.
//
// It lives here (not in nativeacc) because the config builders are in the
// generated service packages, which nativeacc transitively imports: embedding a
// nativeacc type would form an import cycle. The interface is structural, so
// satisfying nativeacc.ResourceConfig needs no import of nativeacc.
type ResourceConfigBase struct {
	tfType string
	label  string
}

// defaultLabel is the Terraform state label a config builder gets when it omits one —
// the common single-instance case. Supply an explicit, meaningful label only when a
// scope holds more than one instance of a type (e.g. "primary"/"secondary") or for a
// validation config that must differ from the resource-under-test ("invalid").
const defaultLabel = "test"

// NewResourceConfigBase builds the embedded base for a config builder's NewXxxCfg
// constructor: tfType is the generated Descriptor's Name. The label is optional —
// omit it for the single-instance default ("test"), or pass exactly one explicit
// label for a multi-instance scope. More than one label is a programming error.
func NewResourceConfigBase(tfType string, label ...string) ResourceConfigBase {
	l := defaultLabel
	switch len(label) {
	case 0:
	case 1:
		l = label[0]
	default:
		panic("generated: NewResourceConfigBase accepts at most one label")
	}
	return ResourceConfigBase{tfType: tfType, label: l}
}

// ResourceType returns the Terraform type (implements nativeacc.ResourceConfig).
func (b ResourceConfigBase) ResourceType() string { return b.tfType }

// ResourceLabel returns the Terraform state label (implements nativeacc.ResourceConfig).
func (b ResourceConfigBase) ResourceLabel() string { return b.label }

// IDRef is the Terraform reference to this resource's id attribute, e.g.
// "azapi_resource_group.rg.id".
func (b ResourceConfigBase) IDRef() string { return b.tfType + "." + b.label + ".id" }

// RefOf is the Terraform reference to this resource's arbitrary attribute, e.g.
// "azapi_resource_group.rg.location" for RefOf("location"). It derives from the same type.label as the acceptance Resource
func (b ResourceConfigBase) RefOf(path string) string { return b.tfType + "." + b.label + "." + path }

type DataSourceConfigBase struct {
	tfType string
	label  string
}

// DataSourceType implements [nativeacc.DataSourceConfig].
func (b DataSourceConfigBase) DataSourceType() string {
	return b.tfType
}

// Label implements [nativeacc.DataSourceConfig].
func (b DataSourceConfigBase) Label() string {
	return b.label
}

func (b DataSourceConfigBase) IDRef() string { return "data." + b.tfType + "." + b.label + ".id" }

// RefOf is the Terraform reference to this resource's arbitrary attribute, e.g.
// "azapi_resource_group.rg.location" for RefOf("location"). It derives from the same type.label as the acceptance Resource
func (b DataSourceConfigBase) RefOf(path string) string {
	return "data." + b.tfType + "." + b.label + "." + path
}

func (r DataSourceConfigBase) IsDataSourceConfig() {}

var (
	ClientConfig = DataSourceConfigBase{tfType: "azapi_client_config", label: "current"}
)
