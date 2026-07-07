package nativeacc

import "github.com/onsi/gomega/format"

// Configure is implemented by a scenario config: a concrete acceptance-test state
// (create baseline, update state, validation case, replacement case, etc.) that
// renders one Terraform block. Resource metadata and scenario HCL stay separate:
// ResourceConfig identifies the Terraform address for ResourceFor, while Configure is
// what Apply/Stage/ApplyExpectError write to disk.
type Configure interface {
	Config() string
}

// StringConfigure adapts one-off literal HCL to Configure. Prefer named scenario
// types for acceptance scenarios; use this only when a test needs raw inline config.
type StringConfigure string

func (s StringConfigure) Config() string { return string(s) }

// ResourceConfig is implemented by a generated resource config builder via the
// embedded services.ResourceConfigBase: it names the Terraform type and state label
// of the resource it configures. A scope's ResourceFor takes one and vends the
// matching Resource handle, so the literal type and label are spelled exactly once —
// set from the resource's generated Descriptor.Name in the builder's NewXxxCfg
// constructor — not restated here. Scenario types wrap this builder and implement
// Configure for the specific HCL state being applied.
type ResourceConfig interface {
	ResourceType() string     // Terraform type, e.g. "azapi_storage_account"
	ResourceLabel() string    // Terraform state label
	IDRef() string            // Terraform reference to the resource's id attribute, e.g. "azapi_resource_group.rg.id".
	RefOf(path string) string // Terraform reference to an arbitrary attribute, e.g. "azapi_resource_group.rg.location" for RefOf("location").
}

// DataSourceConfig is implemented by a data-source config builder (e.g. the shared
// azapi_client_config): it names the Terraform type and label and renders the data
// block. Scope.DataSource writes Config() to disk once; dependents reference the data
// source by address via RefOf/IDRef rather than redeclaring the block.
type DataSourceConfig interface {
	DataSourceType() string   // Terraform type, e.g. "azapi_client_config"
	Label() string            // Terraform state label
	Config() string           // the rendered data block, e.g. `data "azapi_client_config" "current" {}`
	IDRef() string            // Terraform reference to the data source's id attribute, e.g. "data.azapi_client_config.current.id".
	RefOf(path string) string // Terraform reference to an arbitrary attribute.
}

func init() {
	format.MaxDepth = 1 // avoid truncating nested structs in gomega diffs
}
