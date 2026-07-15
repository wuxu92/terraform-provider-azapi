package kusto

import (
	"github.com/Azure/terraform-provider-azapi/internal/native/services/config"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
)

// KustoClusterCfg carries the Terraform address metadata and resource-group
// dependency for azapi_kusto_cluster acceptance-test scenarios. Construct it with
// NewKustoClusterCfg, then wrap it in a scenario type when applying. The parent
// ResourceGroupCfg is held so every scenario renders the same resource_group_id
// reference (e.g. "azapi_resource_group.rg.id"), and so the Kusto database scenarios
// can reference this cluster's IDRef as their cluster_id. The returned HCL is a
// template rendered by the acceptance framework ({{.RandomString}}, {{.Location}}).
type KustoClusterCfg struct {
	config.ResourceConfigBase
	resourceGroup resources.ResourceGroupCfg
}

// NewKustoClusterCfg builds a Kusto-cluster config depending on the parent
// resource-group config: the cluster holds it and references its IDRef as
// resource_group_id. The label is optional — omit it for the single-instance default
// ("test"), or pass an explicit label when a scope holds more than one. The resource
// type is read from the KustoCluster descriptor.
func NewKustoClusterCfg(resourceGroup resources.ResourceGroupCfg, label ...string) KustoClusterCfg {
	return KustoClusterCfg{
		ResourceConfigBase: config.NewResourceConfigBase(KustoCluster.Name, label...),
		resourceGroup:      resourceGroup,
	}
}

// KustoClusterCfg_Basic is a minimal single-node dev-tier cluster: the cheapest valid
// SKU (Dev(No SLA)_Standard_E2a_v4 / Basic / capacity 1). The empty properties block
// is deliberate — it makes properties a present (non-null) object so the framework
// descends into it and applies any nested computed defaults.
type KustoClusterCfg_Basic KustoClusterCfg

func (r KustoClusterCfg_Basic) Config() string {
	return KustoClusterCfg(r).config(`
  sku = {
    name     = "Dev(No SLA)_Standard_E2a_v4"
    tier     = "Basic"
    capacity = 1
  }
  properties = {}`)
}

func (r KustoClusterCfg) config(body string) string {
	return r.RenderConfig(config.ConfigEnvelope{
		// Kusto cluster names are globally unique, lowercase alphanumeric, 4-22 chars.
		Name:       "accazapiadx" + "{{.RandomString}}",
		ParentAttr: "resource_group_id",
		ParentRef:  r.resourceGroup.IDRef(),
		Location:   true,
		Body:       body,
	})
}
