package documentdb_test

import (
	. "github.com/onsi/ginkgo/v2"

	acc "github.com/Azure/terraform-provider-azapi/internal/native/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/documentdb"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
)

// Azure Cosmos DB exercises a native resource with a nested discriminated property:
// properties.backup_policy is discriminated by type (Periodic / Continuous). The
// resource group base lives at the workspace root scope; the account scenario creates
// a minimal single-region GlobalDocumentDB account (no backup_policy → Azure's default
// Periodic), imports it drift-free, then sets the Periodic variant explicitly and
// migrates to the Continuous variant in place — each Apply proving the generated
// AtMostOneOf over the variant blocks accepts exactly the configured one.
var _ = Describe("Azure Cosmos DB Account", Ordered, func() {
	ws := acc.NewWorkspace()
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)

	// Root scope: the resource group base reused by the account scenario.
	rgCfg := resources.NewResourceGroupCfg()
	rg := ws.ResourceFor(rgCfg)
	BeforeAll(func() {
		rg.Apply(resources.ResourceGroupCfg_Basic(rgCfg), acc.Exists())
	})

	cfg := documentdb.NewDocumentDBDatabaseAccountCfg(rgCfg)
	account := ws.ResourceFor(cfg)

	It("creates a Cosmos DB account, then imports it with no drift", func() {
		account.Apply(documentdb.DocumentDBDatabaseAccountCfg_Basic(cfg), acc.Exists()).ImportVerify()
	})

	When("backup policy", func() {
		It("sets the Periodic variant in place", func() {
			// The Periodic variant block populates the discriminated backup_policy; each
			// Apply re-plans for drift, so the in-place change round-trips.
			account.Apply(documentdb.DocumentDBDatabaseAccountCfg_PeriodicBackup(cfg)).ImportVerify()
		})

		It("migrates to the Continuous variant in place", func() {
			// Swapping to the Continuous variant block proves the AtMostOneOf accepts the
			// other variant and the mapper flattens the newly-active block, nulling the
			// inactive one (Periodic -> Continuous is a one-way in-place migration).
			account.Apply(documentdb.DocumentDBDatabaseAccountCfg_ContinuousBackup(cfg)).ImportVerify()
		})
	})
})
