package managedidentity

import (
	"github.com/Azure/terraform-provider-azapi/internal/native/services/config"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
)

// UserAssignedIdentityCfg carries the Terraform address metadata and resource-group
// dependency for azapi_user_assigned_identity acceptance-test
// scenarios. Construct it with NewUserAssignedIdentityCfg, then wrap it in a scenario
// type when applying. The parent ResourceGroupCfg is held so every scenario renders
// the same resource_group_id reference (e.g. "azapi_resource_group.rg.id"). The
// returned HCL is a template rendered by the acceptance framework ({{.RandomString}},
// {{.Location}}).
type UserAssignedIdentityCfg struct {
	config.ResourceConfigBase
	resourceGroup resources.ResourceGroupCfg
}

// NewUserAssignedIdentityCfg builds a user-assigned-identity config depending on the
// parent resource-group config (e.g. the one applied as the base): the identity holds
// it and references its IDRef as resource_group_id. The label is optional — omit it for
// the single-instance default ("test"), or pass an explicit label when a scope holds
// more than one. The resource type is read from the UserAssignedIdentity
// descriptor.
func NewUserAssignedIdentityCfg(resourceGroup resources.ResourceGroupCfg, label ...string) UserAssignedIdentityCfg {
	return UserAssignedIdentityCfg{
		ResourceConfigBase: config.NewResourceConfigBase(UserAssignedIdentity.Name, label...),
		resourceGroup:      resourceGroup,
	}
}

// UserAssignedIdentityCfg_Basic is a minimal identity: name, resource_group_id and
// location only. The properties block is omitted, so ARM applies its default isolation
// scope (None) server-side and the Optional+Computed properties attribute round-trips
// without drift.
type UserAssignedIdentityCfg_Basic UserAssignedIdentityCfg

func (r UserAssignedIdentityCfg_Basic) Config() string {
	return UserAssignedIdentityCfg(r).config("")
}

// UserAssignedIdentityCfg_Complete adds tags — the identity's only in-place-updatable
// user surface (location is ForceNew, the identity coordinates are read-only, and
// isolation_scope's Regional value needs limited-availability subscription support). It
// exercises the create-with-tags path and, applied after Basic, an in-place tag add.
type UserAssignedIdentityCfg_Complete UserAssignedIdentityCfg

func (r UserAssignedIdentityCfg_Complete) Config() string {
	return UserAssignedIdentityCfg(r).config("\n  tags = {\n    environment = \"test\"\n  }")
}

// UserAssignedIdentityCfg_Complete_update flips the tag value set by Complete, so
// applying Complete then Complete_update proves tags survive an in-place update
// (Update -> Read -> empty plan).
type UserAssignedIdentityCfg_Complete_update UserAssignedIdentityCfg

func (r UserAssignedIdentityCfg_Complete_update) Config() string {
	return UserAssignedIdentityCfg(r).config("\n  tags = {\n    environment = \"prod\"\n  }")
}

func (r UserAssignedIdentityCfg) config(body string) string {
	return r.RenderConfig(config.ConfigEnvelope{Name: "acctestuai{{.RandomString}}",
		ParentAttr: "resource_group_id",
		ParentRef:  r.resourceGroup.IDRef(),
		Location:   true,
		Body:       body})
}
