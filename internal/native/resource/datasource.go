package resource

import (
	"context"
	"strings"
	"time"

	"github.com/Azure/terraform-provider-azapi/internal/clients"
	"github.com/Azure/terraform-provider-azapi/internal/native/armjson"
	"github.com/Azure/terraform-provider-azapi/internal/native/mapper"
	"github.com/Azure/terraform-provider-azapi/internal/native/services"
	"github.com/Azure/terraform-provider-azapi/internal/services/parse"
	"github.com/Azure/terraform-provider-azapi/utils"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// DataSource is the generic data source shared by every native static resource.
// It reuses the resource runtime's read path (compose ARM ID, GET, mapper
// flatten) against a schema converted from the generated resource schema, so a
// data source exists for each registered resource without extra codegen.
type DataSource struct {
	desc     services.Descriptor
	provider *clients.Client
}

var (
	_ dschema.DataSource              = &DataSource{}
	_ dschema.DataSourceWithConfigure = &DataSource{}
)

// NewDataSource builds a data source for the given generated Terraform name. The
// provider constructs one per entry in services.Registry, mirroring New.
func NewDataSource(name string) dschema.DataSource {
	d, ok := services.Registry[name]
	if !ok {
		// Programmer error: the provider iterates the registry.
		panic("native: no generated descriptor for " + name)
	}
	return &DataSource{desc: d}
}

func (d *DataSource) typeAndVersion() string { return d.desc.ARMType + "@" + d.desc.APIVersion }

func (d *DataSource) Metadata(_ context.Context, req dschema.MetadataRequest, resp *dschema.MetadataResponse) {
	// desc.Name already carries the provider prefix ("azapi_storage_account").
	resp.TypeName = req.ProviderTypeName + strings.TrimPrefix(d.desc.Name, "azapi")
}

func (d *DataSource) Configure(_ context.Context, req dschema.ConfigureRequest, _ *dschema.ConfigureResponse) {
	if v, ok := req.ProviderData.(*clients.Client); ok {
		d.provider = v
	}
}

func (d *DataSource) Schema(ctx context.Context, _ dschema.SchemaRequest, resp *dschema.SchemaResponse) {
	resp.Schema = d.composeSchema(ctx)
}

// composeSchema converts the generated resource schema into a data-source schema
// (name + parent as Required inputs, everything else Computed) and adds the
// read-only timeouts block — the sole runtime addition, mirroring the resource
// Base (the block needs a context the static generated function cannot hold).
func (d *DataSource) composeSchema(ctx context.Context) schema.Schema {
	s := toDataSourceSchema(d.desc.Schema(), d.desc.ParentAttr)
	if s.Blocks == nil {
		s.Blocks = map[string]schema.Block{}
	}
	s.Blocks["timeouts"] = timeouts.Block(ctx)
	return s
}

func (d *DataSource) objectType(ctx context.Context) basetypes.ObjectType {
	return d.composeSchema(ctx).Type().(basetypes.ObjectType)
}

func (d *DataSource) Read(ctx context.Context, req dschema.ReadRequest, resp *dschema.ReadResponse) {
	if d.provider == nil {
		resp.Diagnostics.AddError("Provider not configured", "the azapi provider was not configured")
		return
	}

	configObj, diags := decodeObject(ctx, d.objectType(ctx), req.Config.Raw)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	bt, err := loadBody(d.desc.ARMType, d.desc.APIVersion)
	if err != nil {
		resp.Diagnostics.AddError("Failed to load resource schema", err.Error())
		return
	}

	id, err := parse.NewResourceID(AttrString(configObj, "name"), AttrString(configObj, d.desc.ParentAttr), d.typeAndVersion())
	if err != nil {
		resp.Diagnostics.AddError("Invalid configuration", err.Error())
		return
	}

	def := d.desc.Timeouts.Read
	if def <= 0 {
		def = 5 * time.Minute
	}
	var to timeouts.Value
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("timeouts"), &to)...)
	if resp.Diagnostics.HasError() {
		return
	}
	timeout, tdiags := to.Read(ctx, def)
	if resp.Diagnostics.Append(tdiags...); resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	respBody, err := d.provider.ResourceClient.Get(ctx, id.AzureResourceId, id.ApiVersion, clients.DefaultRequestOptions())
	if err != nil {
		if utils.ResponseErrorWasNotFound(err) {
			resp.Diagnostics.AddError("Resource not found", "resource "+id.ID()+" was not found")
			return
		}
		resp.Diagnostics.AddError("Failed to retrieve resource", err.Error())
		return
	}

	// Flatten over the config object so practitioner-supplied identity inputs and
	// the timeouts block are preserved while computed body fields are populated
	// from the response.
	newState, fdiags := mapper.FlattenInto(ctx, armjson.AsMap(respBody), configObj, bt)
	if resp.Diagnostics.Append(fdiags...); resp.Diagnostics.HasError() {
		return
	}
	envelope := map[string]string{
		"id":              id.ID(),
		"name":            id.Name,
		d.desc.ParentAttr: id.ParentId,
	}
	for name, val := range envelope {
		newState, diags = withString(ctx, newState, name, val)
		if resp.Diagnostics.Append(diags...); resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(setState(ctx, &resp.State, newState)...)
}
