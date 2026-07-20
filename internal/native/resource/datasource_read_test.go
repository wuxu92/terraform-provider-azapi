package resource

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Azure/terraform-provider-azapi/internal/services/parse"
	dstimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

const (
	dsSubID = "/subscriptions/00000000-0000-0000-0000-000000000000"
	dsName  = "example"
	// dsRGArmID is the ARM id Read composes from the identity inputs
	// (subscription_id + name) — the value the data source must write to state.id.
	dsRGArmID = dsSubID + "/resourceGroups/" + dsName
	// dsRGBody is a valid resource-group GET response: location + a computed
	// provisioningState so mapper.FlattenInto populates the schema without error.
	dsRGBody = `{"id":"` + dsRGArmID + `","name":"` + dsName + `","location":"westeurope","properties":{"provisioningState":"Succeeded"}}`
)

// rgDataConfig builds a tfsdk.Config for the resource-group data source the way
// resourceGroupReadState builds a tfsdk.State: every attribute null except the
// identity inputs Read decodes (name + subscription_id). readTimeout, when
// non-empty, populates timeouts.read so the round-trip case can be exercised.
func rgDataConfig(ctx context.Context, d *DataSource, name, sub, readTimeout string) tfsdk.Config {
	tfType := d.objectType(ctx).TerraformType(ctx).(tftypes.Object)
	vals := make(map[string]tftypes.Value, len(tfType.AttributeTypes))
	for n, at := range tfType.AttributeTypes {
		vals[n] = tftypes.NewValue(at, nil) // null
	}
	vals["name"] = tftypes.NewValue(tftypes.String, name)
	vals["subscription_id"] = tftypes.NewValue(tftypes.String, sub)
	if readTimeout != "" {
		toType := tfType.AttributeTypes["timeouts"].(tftypes.Object)
		vals["timeouts"] = tftypes.NewValue(toType, map[string]tftypes.Value{
			"read": tftypes.NewValue(tftypes.String, readTimeout),
		})
	}
	return tfsdk.Config{
		Raw:    tftypes.NewValue(tfType, vals),
		Schema: d.composeSchema(ctx),
	}
}

func newRGDataSource(t *testing.T, rt *recordingTransport) *DataSource {
	t.Helper()
	ctx := context.Background()
	d := NewDataSource("azapi_resource_group").(*DataSource)
	d.Configure(ctx, datasource.ConfigureRequest{ProviderData: recordingClient(t, rt)}, &datasource.ConfigureResponse{})
	return d
}

func stateString(t *testing.T, st tfsdk.State, p path.Path) string {
	t.Helper()
	ctx := context.Background()
	var v types.String
	if diags := st.GetAttribute(ctx, p, &v); diags.HasError() {
		t.Fatalf("GetAttribute %s: %v", p, diags.Errors())
	}
	return v.ValueString()
}

// TestDataSourceReadHappyPath drives (*DataSource).Read end to end against a
// 200 + valid RG JSON. It proves the read path (a) issues exactly one ARM GET,
// (b) composes state.id from the identity inputs, (c) echoes the configured
// name, and (d) flattens the response location into computed output — the whole
// observable contract of a successful data-source read.
func TestDataSourceReadHappyPath(t *testing.T) {
	ctx := context.Background()
	rt := &recordingTransport{status: http.StatusOK, body: dsRGBody}
	d := newRGDataSource(t, rt)

	req := datasource.ReadRequest{Config: rgDataConfig(ctx, d, dsName, dsSubID, "")}
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: d.composeSchema(ctx)}}
	d.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Read returned error diags: %v", resp.Diagnostics.Errors())
	}
	if rt.calls != 1 {
		t.Fatalf("ARM GET transport was called %d time(s); a successful read must issue exactly one GET", rt.calls)
	}

	wantID, err := parse.NewResourceID(dsName, dsSubID, "Microsoft.Resources/resourceGroups@"+d.desc.APIVersion)
	if err != nil {
		t.Fatalf("compose expected id: %v", err)
	}
	if got := stateString(t, resp.State, path.Root("id")); got != wantID.ID() {
		t.Errorf("state id = %q, want composed ARM id %q", got, wantID.ID())
	}
	if got := stateString(t, resp.State, path.Root("id")); got != dsRGArmID {
		t.Errorf("state id = %q, want %q", got, dsRGArmID)
	}
	if got := stateString(t, resp.State, path.Root("name")); got != dsName {
		t.Errorf("state name = %q, want configured name %q", got, dsName)
	}
	if got := stateString(t, resp.State, path.Root("location")); got != "westeurope" {
		t.Errorf("state location = %q, want %q flattened from the response", got, "westeurope")
	}
}

// TestDataSourceReadNotFound proves the 404 branch: a missing resource surfaces
// the "Resource not found" error diagnostic (not the generic retrieval error),
// and the GET was actually attempted. State is left unset on the error path.
func TestDataSourceReadNotFound(t *testing.T) {
	ctx := context.Background()
	rt := &recordingTransport{
		status: http.StatusNotFound,
		body:   `{"error":{"code":"ResourceGroupNotFound","message":"Resource group 'example' could not be found."}}`,
	}
	d := newRGDataSource(t, rt)

	req := datasource.ReadRequest{Config: rgDataConfig(ctx, d, dsName, dsSubID, "")}
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: d.composeSchema(ctx)}}
	d.Read(ctx, req, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatalf("a 404 GET must produce an error diagnostic; got none")
	}
	if !hasErrorSummary(resp.Diagnostics, "Resource not found") {
		t.Fatalf("want 'Resource not found' diagnostic on 404, got %v", resp.Diagnostics.Errors())
	}
	if rt.calls != 1 {
		t.Fatalf("ARM GET transport was called %d time(s); the not-found branch must follow a real GET", rt.calls)
	}
}

// TestDataSourceReadProviderNotConfigured proves Read guards against an
// unconfigured provider: a DataSource that never received Configure has no
// client and must fail fast with "Provider not configured" rather than
// nil-panic on d.provider.ResourceClient.
func TestDataSourceReadProviderNotConfigured(t *testing.T) {
	ctx := context.Background()
	d := NewDataSource("azapi_resource_group").(*DataSource) // no Configure

	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: d.composeSchema(ctx)}}
	d.Read(ctx, datasource.ReadRequest{}, resp)

	if !hasErrorSummary(resp.Diagnostics, "Provider not configured") {
		t.Fatalf("want 'Provider not configured' diagnostic, got %v", resp.Diagnostics.Errors())
	}
}

// TestDataSourceReadPreservesTimeouts proves the timeouts block round-trips: a
// practitioner-supplied timeouts.read survives FlattenInto (which must not null
// envelope/timeouts values) and lands in the final state unchanged.
func TestDataSourceReadPreservesTimeouts(t *testing.T) {
	ctx := context.Background()
	rt := &recordingTransport{status: http.StatusOK, body: dsRGBody}
	d := newRGDataSource(t, rt)

	req := datasource.ReadRequest{Config: rgDataConfig(ctx, d, dsName, dsSubID, "10m")}
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: d.composeSchema(ctx)}}
	d.Read(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Read returned error diags: %v", resp.Diagnostics.Errors())
	}
	if got := stateString(t, resp.State, path.Root("timeouts").AtName("read")); got != "10m" {
		t.Errorf("state timeouts.read = %q, want configured %q (FlattenInto must not null it)", got, "10m")
	}
	// The preserved value must resolve as the read timeout, not the 5m default.
	var to dstimeouts.Value
	if diags := resp.State.GetAttribute(ctx, path.Root("timeouts"), &to); diags.HasError() {
		t.Fatalf("GetAttribute timeouts: %v", diags.Errors())
	}
	dur, diags := to.Read(ctx, 5*time.Minute)
	if diags.HasError() {
		t.Fatalf("resolve read timeout: %v", diags.Errors())
	}
	if dur != 10*time.Minute {
		t.Errorf("resolved read timeout = %s, want 10m", dur)
	}
}
