package kusto

import (
	"fmt"

	"github.com/Azure/terraform-provider-azapi/internal/native/services/config"
)

// KustoClusterDatabaseCfg carries the Terraform address metadata and cluster
// dependency for azapi_kusto_database acceptance-test scenarios. The database
// body is a discriminated ROOT (kind: ReadWrite / ReadOnlyFollowing): the generated
// schema surfaces one variant block per value at the top level (read_write /
// read_only_following) under an ExactlyOneOf constraint, and the discriminator kind
// is synthesized from whichever block is set. Construct this with
// NewKustoClusterDatabaseCfg, then wrap it in a scenario type when applying. The
// parent KustoClusterCfg is held so every scenario references the same cluster_id.
type KustoClusterDatabaseCfg struct {
	config.ResourceConfigBase
	cluster KustoClusterCfg
}

// NewKustoClusterDatabaseCfg builds a database config depending on the parent
// Kusto-cluster config: the database holds it and references its IDRef as cluster_id.
// The label is optional — omit it for the single-instance default ("test"), or pass
// an explicit label when a scope holds more than one. The resource type is read from
// the KustoClusterDatabase descriptor.
func NewKustoClusterDatabaseCfg(cluster KustoClusterCfg, label ...string) KustoClusterDatabaseCfg {
	return KustoClusterDatabaseCfg{
		ResourceConfigBase: config.NewResourceConfigBase(KustoDatabase.Name, label...),
		cluster:            cluster,
	}
}

// KustoClusterDatabaseCfg_Basic selects the ReadWrite variant of the discriminated
// root — the common read-write database — satisfying the ExactlyOneOf over the
// variant blocks. Both properties (soft/hot cache periods) are ISO-8601 durations.
type KustoClusterDatabaseCfg_Basic KustoClusterDatabaseCfg

func (r KustoClusterDatabaseCfg_Basic) Config() string {
	return KustoClusterDatabaseCfg(r).readWrite("P31D", "P1D")
}

// KustoClusterDatabaseCfg_Update mutates the two in-place-updatable ReadWrite
// properties (soft-delete and hot-cache periods) so applying Basic then Update proves
// the discriminated-root variant survives an in-place Update -> Read -> empty plan.
type KustoClusterDatabaseCfg_Update KustoClusterDatabaseCfg

func (r KustoClusterDatabaseCfg_Update) Config() string {
	return KustoClusterDatabaseCfg(r).readWrite("P60D", "P7D")
}

// readWrite renders the ReadWrite variant block. Defined as a method so the name
// stays scoped to KustoClusterDatabaseCfg and cannot collide with another resource's
// helpers in this shared package.
func (r KustoClusterDatabaseCfg) readWrite(softDelete, hotCache string) string {
	return r.config(fmt.Sprintf(`
  read_write = {
    properties = {
      soft_delete_period = %q
      hot_cache_period   = %q
    }
  }`, softDelete, hotCache))
}

func (r KustoClusterDatabaseCfg) config(body string) string {
	return r.RenderConfig(config.ConfigEnvelope{
		Name:       "accazapidb" + "{{.RandomString}}",
		ParentAttr: "cluster_id",
		Location:   true,
		ParentRef:  r.cluster.IDRef(),
		Body:       body,
	})
}
