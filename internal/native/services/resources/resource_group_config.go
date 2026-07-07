package resources

import (
	"github.com/Azure/terraform-provider-azapi/internal/native/services"
)

// ResourceGroupCfg carries the Terraform address metadata for azapi_resource_group
// acceptance-test scenarios. Construct it with NewResourceGroupCfg, then wrap it in a
// scenario type (ResourceGroupCfg_Basic, ResourceGroupCfg_Named) when applying. The
// returned HCL is a template — {{.RandomInteger}}, {{.Location}} and
// {{.SubscriptionID}} are filled by the acceptance framework's renderer.
type ResourceGroupCfg struct {
	services.ResourceConfigBase
}

// NewResourceGroupCfg builds a resource-group config. The label is optional — omit it
// for the single-instance default ("test"), or pass an explicit label when a scope
// holds more than one. The resource type is read from the ResourceGroup descriptor.
func NewResourceGroupCfg(label ...string) ResourceGroupCfg {
	return ResourceGroupCfg{services.NewResourceConfigBase(ResourceGroup.Name, label...)}
}

// ResourceGroupCfg_Basic is a minimal resource group named acctest-rg-<n>.
type ResourceGroupCfg_Basic ResourceGroupCfg

func (r ResourceGroupCfg_Basic) Config() string {
	return ResourceGroupCfg(r).config("accazapi-rg-{{.RandomInteger}}")
}

// ResourceGroupCfg_Named renders a resource group with an explicit name, for
// validation scenarios (e.g. asserting a name that violates the schema regex is
// rejected at plan time).
type ResourceGroupCfg_Named struct {
	ResourceGroupCfg
	Name string
}

func (r ResourceGroupCfg_Named) Config() string {
	return r.config(r.Name)
}

func (r ResourceGroupCfg) config(name string) string {
	return r.RenderConfig(services.ConfigEnvelope{
		Name:       name,
		ParentAttr: "subscription_id",
		ParentRef:  `"/subscriptions/{{.SubscriptionID}}"`,
		Location:   true,
	})
}
