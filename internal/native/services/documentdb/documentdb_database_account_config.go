package documentdb

import (
	"github.com/Azure/terraform-provider-azapi/internal/native/services/config"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
)

// DocumentDBDatabaseAccountCfg carries the Terraform address metadata and
// resource-group dependency for azapi_cosmosdb_account acceptance-test
// scenarios. The account body carries a nested discriminated property,
// properties.backup_policy (discriminated by type: Periodic / Continuous): the
// generated schema surfaces periodic/continuous variant blocks under an AtMostOneOf
// constraint, with the discriminator type synthesized from whichever block is set.
// Construct this with NewDocumentDBDatabaseAccountCfg, then wrap it in a scenario type
// when applying. The parent ResourceGroupCfg is held so every scenario renders the
// same resource_group_id reference. The returned HCL is a template rendered by the
// acceptance framework ({{.RandomString}}, {{.Location}}).
type DocumentDBDatabaseAccountCfg struct {
	config.ResourceConfigBase
	resourceGroup resources.ResourceGroupCfg
}

// NewDocumentDBDatabaseAccountCfg builds a Cosmos DB account config depending on the
// parent resource-group config: the account holds it and references its IDRef as
// resource_group_id. The label is optional — omit it for the single-instance default
// ("test"), or pass an explicit label when a scope holds more than one. The resource
// type is read from the DocumentdbDatabaseAccount descriptor.
func NewDocumentDBDatabaseAccountCfg(resourceGroup resources.ResourceGroupCfg, label ...string) DocumentDBDatabaseAccountCfg {
	return DocumentDBDatabaseAccountCfg{
		ResourceConfigBase: config.NewResourceConfigBase(CosmosdbAccount.Name, label...),
		resourceGroup:      resourceGroup,
	}
}

// DocumentDBDatabaseAccountCfg_Basic is a minimal GlobalDocumentDB account with the
// required single write region and Standard offer type, and no backup_policy block —
// AtMostOneOf permits zero, so the account rides Azure's default (Periodic) policy.
type DocumentDBDatabaseAccountCfg_Basic DocumentDBDatabaseAccountCfg

func (r DocumentDBDatabaseAccountCfg_Basic) Config() string {
	return DocumentDBDatabaseAccountCfg(r).config("")
}

// DocumentDBDatabaseAccountCfg_PeriodicBackup selects the Periodic variant of the
// nested backup_policy discriminator, exercising the periodic_mode_properties block.
type DocumentDBDatabaseAccountCfg_PeriodicBackup DocumentDBDatabaseAccountCfg

func (r DocumentDBDatabaseAccountCfg_PeriodicBackup) Config() string {
	return DocumentDBDatabaseAccountCfg(r).config(`
    backup_policy = {
      periodic = {
        periodic_mode_properties = {
          backup_interval_in_minutes         = 240
          backup_retention_interval_in_hours = 8
          backup_storage_redundancy          = "Local"
        }
      }
    }`)
}

// DocumentDBDatabaseAccountCfg_ContinuousBackup selects the Continuous variant of the
// nested backup_policy discriminator, exercising the continuous_mode_properties block
// and proving the AtMostOneOf swaps cleanly from the Periodic variant. Migrating from
// Periodic to Continuous is a one-way in-place change Azure supports.
type DocumentDBDatabaseAccountCfg_ContinuousBackup DocumentDBDatabaseAccountCfg

func (r DocumentDBDatabaseAccountCfg_ContinuousBackup) Config() string {
	return DocumentDBDatabaseAccountCfg(r).config(`
    backup_policy = {
      continuous = {
        continuous_mode_properties = {
          tier = "Continuous7Days"
        }
      }
    }`)
}

// config renders the account with the required single-region GlobalDocumentDB body,
// appending the given backup_policy fragment (empty for Basic) inside properties.
func (r DocumentDBDatabaseAccountCfg) config(backupPolicy string) string {
	body := `
  properties = {
    database_account_offer_type = "Standard"
    locations = [
      {
        location_name     = "{{.Location}}"
        failover_priority = 0
      }
    ]` + backupPolicy + `
  }`
	return r.RenderConfig(config.ConfigEnvelope{
		// Cosmos DB account names are globally unique, lowercase alphanumeric + hyphen.
		Name:       "accazapicosmos" + "{{.RandomString}}",
		ParentAttr: "resource_group_id",
		ParentRef:  r.resourceGroup.IDRef(),
		Location:   true,
		Kind:       "GlobalDocumentDB",
		Body:       body,
	})
}
