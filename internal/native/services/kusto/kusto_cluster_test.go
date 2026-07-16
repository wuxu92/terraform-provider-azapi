package kusto_test

import (
	. "github.com/onsi/ginkgo/v2"

	acc "github.com/Azure/terraform-provider-azapi/internal/native/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/kusto"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
)

// Azure Data Explorer (Kusto) cluster exercises the standalone lifecycle of the
// parent resource: a minimal single-node dev-tier cluster is created and imported
// drift-free, then a broad in-place update adds tags plus the mutable scalar surface
// (streaming ingest, purge, public network access), and a second update flips every
// one of those values — each Apply re-plans for drift, so the round-trips prove the
// update path and the mapper flatten without per-value assertions. The Required SKU
// block is held stable across all three (SKU scaling is a separate concern).
var _ = Describe("Azure Data Explorer Cluster", Ordered, func() {
	ws := acc.NewWorkspace()
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)

	// Root scope: the resource group base reused by the cluster scenario.
	rgCfg := resources.NewResourceGroupCfg("rg")
	rg := ws.ResourceFor(rgCfg)
	BeforeAll(func() { rg.Apply(resources.ResourceGroupCfg_Basic(rgCfg), acc.Exists()) })

	cfg := kusto.NewKustoClusterCfg(rgCfg, "adx")
	cluster := ws.ResourceFor(cfg)

	It("creates a dev-tier cluster, then imports it with no drift", func() {
		cluster.Apply(kusto.KustoClusterCfg_Basic(cfg), acc.Exists()).ImportVerify()
	})

	It("adds the mutable cluster surface in place", func() {
		// Complete adds tags plus streaming-ingest/purge/public-network scalars over
		// Basic; each Apply re-plans for drift, so the in-place update round-trips.
		cluster.Apply(kusto.KustoClusterCfg_Complete(cfg), acc.Exists()).ImportVerify()
	})

	It("re-updates every mutable cluster value in place", func() {
		// Complete_update flips each value set by Complete to a different valid value;
		// the round-trip proves each survives an in-place Update -> Read -> empty plan.
		cluster.Apply(kusto.KustoClusterCfg_Complete_update(cfg)).ImportVerify()
	})
})
