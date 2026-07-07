// Package config holds the acceptance-test config scaffolding shared by every
// generated service package's *Cfg builder: the resource/data-source base types,
// the envelope renderer, and the shared azapi_client_config data source. It is a
// dedicated package (not the parent services package, whose registry.go is the
// runtime descriptor registry the generated _gen.go files register into) so the
// test-only config helpers and the runtime registry stay cleanly separated. Service
// subpackages import this for their config builders; nothing here imports them.
package config

import (
	"fmt"
	"strings"
)

// ResourceConfigBase carries the Terraform type and state label shared by every
// acceptance-test config builder. A builder embeds it (constructing it via
// NewResourceConfigBase in its NewXxxCfg constructor) to get ResourceType and
// ResourceLabel for free and thereby satisfy nativeacc.ResourceConfig, so a
// scope's ResourceFor can vend the matching Resource handle. The type is set
// once, from the resource's generated Descriptor.Name — callers never restate it.
//
// It lives in this dedicated package (not in nativeacc) because the config
// builders are in the generated service packages, which nativeacc transitively
// imports: embedding a nativeacc type would form an import cycle. The interface is
// structural, so satisfying nativeacc.ResourceConfig needs no import of nativeacc.
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

// ConfigEnvelope is the set of envelope-varying inputs a static resource's
// acceptance config builder hands to ResourceConfigBase.RenderConfig. It captures
// exactly the axes along which the shared `resource … { name / <parent> / location /
// kind }` shell differs between the generated service packages' *Cfg builders; the
// irreducible, semantics-bearing property body stays in each builder and arrives here
// as Body, appended verbatim. Type and label come from the embedded ResourceConfigBase.
type ConfigEnvelope struct {
	// Name is the resource's name attribute value; the renderer quotes it, so pass the
	// raw (possibly templated) value, e.g. "acctestsa{{.RandomString}}" or "default".
	Name string
	// ParentAttr is the parent-reference attribute name: "subscription_id",
	// "resource_group_id" or "storage_account_id". It is always the widest envelope key,
	// so the renderer aligns name/location/kind to its width.
	ParentAttr string
	// ParentRef is the parent-reference right-hand-side expression, emitted verbatim: a
	// quoted subscription path (`"/subscriptions/{{.SubscriptionID}}"`) or a dependency's
	// IDRef() (e.g. azapi_resource_group.test.id).
	ParentRef string
	// Location emits `location = "{{.Location}}"`; false only for the singleton blob
	// service child, which has no location of its own.
	Location bool
	// Kind is the kind attribute value ("StorageV2", "app", …); "" omits the kind line.
	Kind string
	// Body is the per-resource property fragment (sku block, properties, …), appended
	// verbatim after the envelope lines. It is empty or begins with a newline; the
	// renderer supplies the closing "\n}\n" and never expands template tokens in it.
	Body string
}

// RenderConfig renders the operational-envelope shell shared by every static
// resource's acceptance config builder: the `resource <type> <label> { name /
// <parent> / location / kind }` scaffolding, with the per-resource Body appended
// verbatim. It is byte-for-byte identical to the shell each builder hand-spelled
// before extraction, so the live acceptance scenarios apply the exact same HCL.
// Template tokens ({{.RandomString}}, {{.Location}}, {{.SubscriptionID}}) and any
// injected IDRef() pass through untouched for the acceptance framework to fill.
func (b ResourceConfigBase) RenderConfig(env ConfigEnvelope) string {
	w := len(env.ParentAttr)
	var sb strings.Builder
	fmt.Fprintf(&sb, "\nresource %q %q {", b.ResourceType(), b.ResourceLabel())
	fmt.Fprintf(&sb, "\n  %-*s = %q", w, "name", env.Name)
	fmt.Fprintf(&sb, "\n  %-*s = %s", w, env.ParentAttr, env.ParentRef)
	if env.Location {
		fmt.Fprintf(&sb, "\n  %-*s = %q", w, "location", "{{.Location}}")
	}
	if env.Kind != "" {
		fmt.Fprintf(&sb, "\n  %-*s = %q", w, "kind", env.Kind)
	}
	sb.WriteString(env.Body)
	sb.WriteString("\n}\n")
	return sb.String()
}

// DataSourceConfigBase carries the Terraform type and state label for an
// acceptance-test data-source config, mirroring ResourceConfigBase for data sources.
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

// RefOf is the Terraform reference to this data source's arbitrary attribute, e.g.
// "data.azapi_client_config.current.tenant_id" for RefOf("tenant_id").
func (b DataSourceConfigBase) RefOf(path string) string {
	return "data." + b.tfType + "." + b.label + "." + path
}

func (r DataSourceConfigBase) IsDataSourceConfig() {}

// Config renders the data-source declaration block, e.g. `data "azapi_client_config"
// "current" {}`.
func (r DataSourceConfigBase) Config() string {
	return fmt.Sprintf(`data %q %q {}`, r.tfType, r.label)
}

// ClientConfigData is the azapi_client_config data source config type.
type ClientConfigData struct {
	DataSourceConfigBase
}

// ClientConfig is the single shared azapi_client_config data source every scenario
// references (via ClientConfig.Config() to declare it and ClientConfig.RefOf(...) for
// its attributes) instead of hardcoding the block. One canonical declaration avoids
// duplicate `data "azapi_client_config" "current"` blocks colliding when more than one
// resource in a workspace needs the running identity's tenant/object/subscription ids.
var ClientConfig = ClientConfigData{DataSourceConfigBase: DataSourceConfigBase{tfType: "azapi_client_config", label: "current"}}
