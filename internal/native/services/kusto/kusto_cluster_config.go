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

// devSKU is the cheapest valid single-node dev-tier SKU block. sku is Required, so
// every scenario's request body must carry it; sharing one fragment keeps the SKU
// stable across create and the in-place updates (a SKU change is a separate scaling
// concern, not exercised here).
const devSKU = `
  sku = {
    name     = "Dev(No SLA)_Standard_E2a_v4"
    tier     = "Basic"
    capacity = 1
  }`

// KustoClusterCfg_Basic is a minimal single-node dev-tier cluster: the cheapest valid
// SKU (Dev(No SLA)_Standard_E2a_v4 / Basic / capacity 1). The empty properties block
// is deliberate — it makes properties a present (non-null) object so the framework
// descends into it and applies any nested computed defaults.
type KustoClusterCfg_Basic KustoClusterCfg

func (r KustoClusterCfg_Basic) Config() string {
	return KustoClusterCfg(r).config(devSKU + `
  properties = {}`)
}

// KustoClusterCfg_Complete adds the cluster's common in-place-updatable surface over
// Basic — tags plus scalar properties (streaming ingest, purge, and public network
// access) — so applying Basic then Complete proves the added surface round-trips
// (Update -> Read -> empty plan). The SKU is unchanged from Basic; ForceNew knobs
// (enable_double_encryption) are held out so the delta is purely mutable-in-place.
type KustoClusterCfg_Complete KustoClusterCfg

func (r KustoClusterCfg_Complete) Config() string {
	return KustoClusterCfg(r).config(devSKU + `
  properties = {
    enable_streaming_ingest = true
    enable_purge            = true
    public_network_access   = "Enabled"
  }
  tags = {
    environment = "test"
  }`)
}

// KustoClusterCfg_Complete_update flips every in-place-updatable value set by Complete
// to a different valid value — retagged, streaming ingest and purge disabled, public
// network access restricted — so applying Complete then Complete_update proves each
// survives an in-place update (Update -> Read -> empty plan).
type KustoClusterCfg_Complete_update KustoClusterCfg

func (r KustoClusterCfg_Complete_update) Config() string {
	return KustoClusterCfg(r).config(devSKU + `
  properties = {
    enable_streaming_ingest = false
    enable_purge            = false
    public_network_access   = "Disabled"
  }
  tags = {
    environment = "prod"
  }`)
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
