package resource

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/terraform-provider-azapi/internal/clients"
	"github.com/Azure/terraform-provider-azapi/internal/native/services"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// fakeCred is a no-network TokenCredential so the ARM pipeline's bearer policy is
// satisfied without contacting Azure; the fakeTransport intercepts every request.
type fakeCred struct{}

func (fakeCred) GetToken(context.Context, policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: "fake", ExpiresOn: time.Now().Add(time.Hour)}, nil
}

// fakeTransport returns a canned response for the existence GET, so Base.put's
// pre-create check can be exercised without a live resource.
type fakeTransport struct {
	status int
	body   string
}

func (f *fakeTransport) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: f.status,
		Status:     http.StatusText(f.status),
		Body:       io.NopCloser(strings.NewReader(f.body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Request:    req,
	}, nil
}

func fakeClient(t *testing.T, status int, body string) *clients.Client {
	t.Helper()
	rc, err := clients.NewResourceClient(fakeCred{}, &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{Transport: &fakeTransport{status: status, body: body}},
	})
	if err != nil {
		t.Fatalf("NewResourceClient: %v", err)
	}
	return &clients.Client{ResourceClient: rc}
}

func stringObject(t *testing.T, kv map[string]string) types.Object {
	t.Helper()
	attrTypes := make(map[string]attr.Type, len(kv))
	attrs := make(map[string]attr.Value, len(kv))
	for k, v := range kv {
		attrTypes[k] = types.StringType
		attrs[k] = types.StringValue(v)
	}
	obj, d := types.ObjectValue(attrTypes, attrs)
	if d.HasError() {
		t.Fatalf("build object: %v", d)
	}
	return obj
}

func hasErrorSummary(diags diag.Diagnostics, summary string) bool {
	for _, d := range diags.Errors() {
		if d.Summary() == summary {
			return true
		}
	}
	return false
}

// TestPutRefusesExistingResourceOnCreate proves Base.put performs the standard
// Terraform pre-create existence check: on create it GETs the target ID first and,
// if the resource already exists, errors out telling the user to import it rather
// than silently overwriting it with the CreateOrUpdate PUT. A non-NotFound GET
// failure is surfaced instead of being swallowed. Both branches return before the
// body is composed, so a minimal name+parent plan object is enough to drive them.
func TestPutRefusesExistingResourceOnCreate(t *testing.T) {
	ctx := context.Background()
	// resource_group: simplest native resource, subscription-scoped. Descriptor is
	// built inline (not via New) so the test needs no registry import, which would
	// cycle back through the generated hook packages into this package.
	desc := services.Descriptor{
		Name:       "azapi_resource_group",
		ARMType:    "Microsoft.Resources/resourceGroups",
		APIVersion: "2025-04-01",
		ParentAttr: "subscription_id",
	}
	planObj := stringObject(t, map[string]string{
		"name":            "acctest-rg",
		"subscription_id": "/subscriptions/00000000-0000-0000-0000-000000000000",
	})

	t.Run("already exists errors with import guidance", func(t *testing.T) {
		b := &Base{desc: desc, provider: fakeClient(t, http.StatusOK, "{}")}
		var diags diag.Diagnostics
		b.put(ctx, planObj, true, timeouts.Value{}, &tfsdk.State{}, &diags)
		if !hasErrorSummary(diags, "Resource already exists") {
			t.Fatalf("want 'Resource already exists' diagnostic, got %v", diags.Errors())
		}
	})

	t.Run("non-not-found GET failure is fatal", func(t *testing.T) {
		b := &Base{desc: desc, provider: fakeClient(t, http.StatusForbidden, `{"error":{"code":"AuthorizationFailed","message":"denied"}}`)}
		var diags diag.Diagnostics
		b.put(ctx, planObj, true, timeouts.Value{}, &tfsdk.State{}, &diags)
		if !hasErrorSummary(diags, "Failed to retrieve resource") {
			t.Fatalf("want 'Failed to retrieve resource' diagnostic, got %v", diags.Errors())
		}
	})
}
