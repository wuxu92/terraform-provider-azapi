package kusto_test

import (
	. "github.com/onsi/ginkgo/v2"

	acc "github.com/Azure/terraform-provider-azapi/internal/native/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/kusto"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
)

// Azure Data Explorer (Kusto) exercises a discriminated-ROOT native resource. A
// resource group (root scope) and a Kusto cluster (child scope) are each provisioned
// once from their own config builders; the database — whose body is a discriminated
// type (kind: ReadWrite / ReadOnlyFollowing) — selects the ReadWrite variant, so the
// generated ExactlyOneOf over the top-level variant blocks is satisfied by exactly one
// block. The database lives in the cluster's scope and tears down with it.
var _ = Describe("Azure Data Explorer Database", Ordered, func() {
	ws := acc.NewWorkspace()
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)

	// Root scope: the resource group base, reused by the cluster scope below.
	rgCfg := resources.NewResourceGroupCfg("rg")
	rg := ws.ResourceFor(rgCfg)
	BeforeAll(func() { rg.Apply(resources.ResourceGroupCfg_Basic(rgCfg), acc.Exists()) })

	Describe("on a cluster", Ordered, func() {
		// Child scope: the cluster plus its database, torn down together after this
		// container while the resource group survives.
		adx := ws.Scope()

		clusterCfg := kusto.NewKustoClusterCfg(rgCfg, "adx")
		cluster := adx.ResourceFor(clusterCfg)
		dbCfg := kusto.NewKustoClusterDatabaseCfg(clusterCfg)
		db := adx.ResourceFor(dbCfg)

		BeforeAll(func() {
			// cluster and db share this scope and tear down together, so provision the
			// dependent chain in ONE terraform apply: Stage bundles each resource's
			// config + checks and ApplyAll writes both files, applies once (Terraform
			// orders db after cluster from the cluster_id reference), asserts a single
			// empty post-apply plan for the set, then runs every staged check.
			adx.ApplyAll(
				cluster.Stage(kusto.KustoClusterCfg_Basic(clusterCfg), acc.Exists()),
				db.Stage(kusto.KustoClusterDatabaseCfg_Basic(dbCfg), acc.Exists()),
			)
		})

		It("imports the ReadWrite database with no drift", func() {
			// ApplyAll already asserted existence and a stable plan for the batch;
			// ImportVerify re-imports the database and confirms the read path reproduces
			// the configured ReadWrite-variant state.
			db.ImportVerify()
		})

		It("updates the discriminated-root variant in place", func() {
			// Update flips the two in-place-updatable ReadWrite properties (soft-delete
			// and hot-cache periods); each Apply re-plans for drift, so the update
			// round-trips without per-value assertions.
			db.Apply(kusto.KustoClusterDatabaseCfg_Update(dbCfg)).ImportVerify()
		})
	})
})
