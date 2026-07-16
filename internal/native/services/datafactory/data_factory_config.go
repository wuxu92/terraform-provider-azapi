package datafactory

import (
	"github.com/Azure/terraform-provider-azapi/internal/native/services/config"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
)

// DataFactoryCfg carries the Terraform address metadata and resource-group
// dependency for azapi_data_factory acceptance-test scenarios. Construct it with
// NewDataFactoryCfg, then wrap it in a scenario type when applying. The parent
// ResourceGroupCfg is held so every scenario renders the same resource_group_id
// reference (e.g. "azapi_resource_group.test.id"). The returned HCL is a template
// rendered by the acceptance framework ({{.RandomString}}, {{.Location}}).
//
// The factory body carries a nested discriminated property,
// properties.repo_configuration (discriminated by type: FactoryGitHubConfiguration
// / FactoryVSTSConfiguration): the generated schema surfaces the two variant blocks
// under an AtMostOneOf constraint. Those variants require a live Git account/PAT to
// apply, so they are intentionally not exercised here; the scenarios below cover the
// dependency-free create/import and an in-place update instead.
type DataFactoryCfg struct {
	config.ResourceConfigBase
	resourceGroup resources.ResourceGroupCfg
}

// NewDataFactoryCfg builds a data-factory config depending on the parent
// resource-group config: the factory holds it and references its IDRef as
// resource_group_id. The label is optional — omit it for the single-instance default
// ("test"), or pass an explicit label when a scope holds more than one. The resource
// type is read from the DataFactory descriptor.
func NewDataFactoryCfg(resourceGroup resources.ResourceGroupCfg, label ...string) DataFactoryCfg {
	return DataFactoryCfg{
		ResourceConfigBase: config.NewResourceConfigBase(DataFactory.Name, label...),
		resourceGroup:      resourceGroup,
	}
}

// DataFactoryCfg_Basic is a minimal factory: name + location + an empty properties
// block. The factory body has no required properties, so the empty (but present)
// properties object lets the framework descend and settle any nested computed
// defaults (provisioning_state, version, public_network_access).
type DataFactoryCfg_Basic DataFactoryCfg

func (r DataFactoryCfg_Basic) Config() string {
	return DataFactoryCfg(r).config(`
  properties = {}`)
}

// DataFactoryCfg_Update flips a single in-place-updatable factory setting —
// disabling public network access — with no external dependency, proving the update
// path round-trips and re-plans drift-free. (public_network_access is a plain enum
// that ARM echoes back verbatim, so the flatten is exact.)
type DataFactoryCfg_Update DataFactoryCfg

func (r DataFactoryCfg_Update) Config() string {
	return DataFactoryCfg(r).config(`
  properties = {
    public_network_access = "Disabled"
  }`)
}

func (r DataFactoryCfg) config(body string) string {
	return r.RenderConfig(config.ConfigEnvelope{
		// Data Factory names are globally unique, 3-63 chars, alphanumeric and hyphens,
		// starting and ending with an alphanumeric character.
		Name:       "acc-azapi-adf-" + "{{.RandomString}}",
		ParentAttr: "resource_group_id",
		ParentRef:  r.resourceGroup.IDRef(),
		Location:   true,
		Body:       body,
	})
}
