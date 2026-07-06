package resource

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/terraform-provider-azapi/internal/clients"
	"github.com/Azure/terraform-provider-azapi/internal/native/services"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// recordingTransport is a fakeTransport that also counts how many times Do was
// invoked, so a test can prove the ARM GET was (or was not) issued.
type recordingTransport struct {
	status int
	body   string
	calls  int
}

func (r *recordingTransport) Do(req *http.Request) (*http.Response, error) {
	r.calls++
	return &http.Response{
		StatusCode: r.status,
		Status:     http.StatusText(r.status),
		Body:       io.NopCloser(strings.NewReader(r.body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Request:    req,
	}, nil
}

// recordingClient builds a real clients.Client whose ResourceClient routes every
// request through rt, letting the caller inspect rt.calls afterwards.
func recordingClient(t *testing.T, rt *recordingTransport) *clients.Client {
	t.Helper()
	rc, err := clients.NewResourceClient(fakeCred{}, &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{Transport: rt},
	})
	if err != nil {
		t.Fatalf("NewResourceClient: %v", err)
	}
	return &clients.Client{ResourceClient: rc}
}

// resourceGroupReadState builds a minimal but fully valid tfsdk.State for Base.Read
// against the resource-group descriptor: a state Raw whose only populated attribute
// is a parseable ARM id (every other attribute, including the timeouts block, is
// null), plus the composed schema so State.GetAttribute(timeouts) resolves. This is
// exactly what Read decodes via rawToObject before it reaches the BeforeRead hook.
func resourceGroupReadState(ctx context.Context, b *Base, armID string) tfsdk.State {
	objType := b.objectType(ctx)
	tfType := objType.TerraformType(ctx).(tftypes.Object)
	vals := make(map[string]tftypes.Value, len(tfType.AttributeTypes))
	for name, at := range tfType.AttributeTypes {
		vals[name] = tftypes.NewValue(at, nil) // null
	}
	vals["id"] = tftypes.NewValue(tftypes.String, armID)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tfType, vals),
		Schema: b.composeSchema(ctx),
	}
}

// TestReadFiresBeforeReadHookBeforeGet proves the load-bearing ordering contract in
// Base.Read: the BeforeRead hook runs BEFORE the ARM GET, and an error diagnostic
// appended by the hook short-circuits Read so the GET is never issued. This drives
// Base.Read end to end (state decode, id parse, timeout resolution, hook, GET gate)
// rather than merely checking hook registration.
func TestReadFiresBeforeReadHookBeforeGet(t *testing.T) {
	ctx := context.Background()
	desc := services.Descriptor{
		Name:       "azapi_resource_group",
		ARMType:    "Microsoft.Resources/resourceGroups",
		APIVersion: "2025-04-01",
		ParentAttr: "subscription_id",
		Schema:     resources.AzapiResourceGroupSchema,
	}
	const armID = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acctest-rg"

	rt := &recordingTransport{status: http.StatusOK, body: "{}"}
	b := &Base{desc: desc, provider: recordingClient(t, rt)}

	hookRan := false
	b.hooks = &Hooks{
		BeforeRead: func(hc *CrudCtx) {
			hookRan = true
			// The hook must observe the state Read decoded and the parsed ID.
			if got := attrString(hc.State, "id"); got != armID {
				t.Errorf("BeforeRead saw state id %q, want %q", got, armID)
			}
			if hc.ID.AzureResourceId != armID {
				t.Errorf("BeforeRead saw parsed ID %q, want %q", hc.ID.AzureResourceId, armID)
			}
			hc.Diags.AddError("BeforeRead blocked", "hook short-circuited the read")
		},
	}

	req := resource.ReadRequest{State: resourceGroupReadState(ctx, b, armID)}
	resp := &resource.ReadResponse{}
	b.Read(ctx, req, resp)

	if !hookRan {
		t.Fatal("BeforeRead hook was never invoked")
	}
	if !hasErrorSummary(resp.Diagnostics, "BeforeRead blocked") {
		t.Fatalf("want 'BeforeRead blocked' diagnostic from the hook, got %v", resp.Diagnostics.Errors())
	}
	if rt.calls != 0 {
		t.Fatalf("ARM GET transport was called %d time(s); BeforeRead must precede and short-circuit the GET", rt.calls)
	}
}
