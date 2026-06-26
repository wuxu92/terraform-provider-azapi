package generated

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

// NewResourceConfigBase builds the embedded base for a config builder's NewXxxCfg
// constructor: tfType is the generated Descriptor's Name, label is the state label.
func NewResourceConfigBase(tfType, label string) ResourceConfigBase {
	return ResourceConfigBase{tfType: tfType, label: label}
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

func (r *ResourceConfigBase) Config() string {
	return "resource " + r.tfType + " " + r.label + " {}"
}

type DataSourceConfigBase struct {
	tfType string
	label  string
}

// DataSourceType implements [nativeacc.DataSourceConfig].
func (b *DataSourceConfigBase) DataSourceType() string {
	return b.tfType
}

// Label implements [nativeacc.DataSourceConfig].
func (b *DataSourceConfigBase) Label() string {
	return b.label
}

func (b DataSourceConfigBase) IDRef() string { return "data." + b.tfType + "." + b.label + ".id" }

// RefOf is the Terraform reference to this resource's arbitrary attribute, e.g.
// "azapi_resource_group.rg.location" for RefOf("location"). It derives from the same type.label as the acceptance Resource
func (b DataSourceConfigBase) RefOf(path string) string {
	return "data." + b.tfType + "." + b.label + "." + path
}

func (r *DataSourceConfigBase) Config() string {
	return "data " + r.tfType + " " + r.label + " {}"
}
func (r *DataSourceConfigBase) IsDataSourceConfig() {}


var (
	ClientConfig = DataSourceConfigBase{tfType: "azapi_client_config", label: "current"}
)