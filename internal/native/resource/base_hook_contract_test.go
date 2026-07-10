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

// This file is the drift guard for the per-hook-point field-liveness contract
// documented on CrudCtx/Hooks (hooks.go). It drives the REAL Base CRUD entry
// points in-process and, inside each hook body, snapshots which CrudCtx fields
// are live vs null at the instant the hook fires — then asserts they match the
// contract matrix. If a future change to Base's orchestration stops populating a
// documented-live field (or starts populating a documented-null one), these
// tests fail rather than shipping a silent contract violation.
//
// It is deliberately NOT a registration check (the *HookRegistered tests cover
// that) and NOT an assertion on unexported call sequencing: it observes only the
// external CrudCtx a hook author actually receives.

// liveness is the field-liveness snapshot of a CrudCtx at one hook point. Ctx,
// Client, ID and Diags are always live, so only the four variable fields are
// recorded.
type liveness struct {
	planLive     bool
	stateLive    bool
	bodyLive     bool
	responseLive bool
}

// snapCrud captures field liveness at the moment a hook fires. It MUST be called
// inside the hook body: Base.put reuses one *CrudCtx pointer across the Before*
// and After* hook of a create/update (it sets hc.Response between them), so a
// pointer stashed and inspected after the call would show the Before* hook a
// Response that was only populated later. Plan/State are types.Object (null via
// IsNull); Body/Response are maps (live means non-nil).
func snapCrud(hc *CrudCtx) liveness {
	return liveness{
		planLive:     !hc.Plan.IsNull(),
		stateLive:    !hc.State.IsNull(),
		bodyLive:     hc.Body != nil,
		responseLive: hc.Response != nil,
	}
}

// scriptTransport answers each request by method + call index, so one test can
// drive the existence-GET -> PUT -> readback-GET sequence of a full create.
type scriptTransport struct {
	n       int
	handler func(req *http.Request, n int) (int, string)
}

func (s *scriptTransport) Do(req *http.Request) (*http.Response, error) {
	s.n++
	status, body := s.handler(req, s.n)
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Request:    req,
	}, nil
}

// scriptClient builds a real clients.Client whose ResourceClient routes every
// request through a scriptTransport driven by h.
func scriptClient(t *testing.T, h func(*http.Request, int) (int, string)) *clients.Client {
	t.Helper()
	rc, err := clients.NewResourceClient(fakeCred{}, &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{Transport: &scriptTransport{handler: h}},
	})
	if err != nil {
		t.Fatalf("NewResourceClient: %v", err)
	}
	return &clients.Client{ResourceClient: rc}
}

// resourceGroupPlanState builds a valid create/update plan (name + subscription_id
// + location set, everything else null) for the resource-group descriptor, plus
// the composed schema so Base.Create/Update can resolve the timeouts block. The
// return type is tfsdk.State because callers convert it to tfsdk.Plan (both are
// {Raw, Schema}); Base decodes it via rawToObject exactly as the framework would.
func resourceGroupPlanState(ctx context.Context, b *Base) tfsdk.State {
	objType := b.objectType(ctx)
	tfType := objType.TerraformType(ctx).(tftypes.Object)
	vals := make(map[string]tftypes.Value, len(tfType.AttributeTypes))
	for name, at := range tfType.AttributeTypes {
		vals[name] = tftypes.NewValue(at, nil) // null
	}
	vals["name"] = tftypes.NewValue(tftypes.String, "acctest-rg")
	vals["subscription_id"] = tftypes.NewValue(tftypes.String, "/subscriptions/00000000-0000-0000-0000-000000000000")
	vals["location"] = tftypes.NewValue(tftypes.String, "westeurope")
	return tfsdk.State{Raw: tftypes.NewValue(tfType, vals), Schema: b.composeSchema(ctx)}
}

// resourceGroupDescriptor is the simplest native driver: subscription-scoped, no
// singleton. Built inline (not via New) so the test needs no registry import,
// which would cycle back through the generated hook packages into this package.
func resourceGroupDescriptor() services.Descriptor {
	return services.Descriptor{
		Name:       "azapi_resource_group",
		ARMType:    "Microsoft.Resources/resourceGroups",
		APIVersion: "2025-04-01",
		ParentAttr: "subscription_id",
		Schema:     resources.AzapiResourceGroupSchema,
	}
}

const rgArmID = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acctest-rg"

// rgBody is an ARM resource-group JSON that flattens cleanly into the schema, so
// the create/update/read paths run to completion without a mapping error.
const rgBody = `{"id":"` + rgArmID + `","location":"westeurope","properties":{"provisioningState":"Succeeded"}}`

// assertLiveness compares an observed snapshot against the contract and reports
// each divergent field, so a failure names exactly which liveness promise broke.
func assertLiveness(t *testing.T, hook string, got, want liveness) {
	t.Helper()
	if got.planLive != want.planLive {
		t.Errorf("%s: Plan live = %v, contract says %v", hook, got.planLive, want.planLive)
	}
	if got.stateLive != want.stateLive {
		t.Errorf("%s: State live = %v, contract says %v", hook, got.stateLive, want.stateLive)
	}
	if got.bodyLive != want.bodyLive {
		t.Errorf("%s: Body live = %v, contract says %v", hook, got.bodyLive, want.bodyLive)
	}
	if got.responseLive != want.responseLive {
		t.Errorf("%s: Response live = %v, contract says %v", hook, got.responseLive, want.responseLive)
	}
}

// TestHookContractFieldLiveness pins the documented CrudCtx field-liveness matrix
// to Base's actual create/update/read/delete orchestration. Each subtest drives a
// real Base entry point end to end and asserts the fields each hook observes.
func TestHookContractFieldLiveness(t *testing.T) {
	ctx := context.Background()

	t.Run("create", func(t *testing.T) {
		desc := resourceGroupDescriptor()
		// create: existence GET (404 -> proceed) -> PUT -> readback GET.
		existenceChecked := false
		b := &Base{desc: desc, provider: scriptClient(t, func(req *http.Request, _ int) (int, string) {
			if req.Method == http.MethodGet && !existenceChecked {
				existenceChecked = true
				return http.StatusNotFound, `{"error":{"code":"ResourceGroupNotFound"}}`
			}
			return http.StatusOK, rgBody
		})}

		var beforeRan, afterRan bool
		var before, after liveness
		b.hooks = &Hooks{
			BeforeCreate: func(hc *CrudCtx) { beforeRan = true; before = snapCrud(hc) },
			AfterCreate:  func(hc *CrudCtx) { afterRan = true; after = snapCrud(hc) },
		}

		plan := resourceGroupPlanState(ctx, b)
		resp := &resource.CreateResponse{}
		b.Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan(plan)}, resp)

		if resp.Diagnostics.HasError() {
			t.Fatalf("Create returned errors: %v", resp.Diagnostics.Errors())
		}
		if !beforeRan {
			t.Fatal("BeforeCreate hook was never invoked")
		}
		if !afterRan {
			t.Fatal("AfterCreate hook was never invoked")
		}
		assertLiveness(t, "BeforeCreate", before, liveness{planLive: true, stateLive: false, bodyLive: true, responseLive: false})
		assertLiveness(t, "AfterCreate", after, liveness{planLive: true, stateLive: false, bodyLive: true, responseLive: true})
	})

	t.Run("update", func(t *testing.T) {
		desc := resourceGroupDescriptor()
		// update: no existence check -> PUT -> readback GET.
		b := &Base{desc: desc, provider: scriptClient(t, func(req *http.Request, _ int) (int, string) {
			return http.StatusOK, rgBody
		})}

		var beforeRan, afterRan bool
		var before, after liveness
		b.hooks = &Hooks{
			BeforeUpdate: func(hc *CrudCtx) { beforeRan = true; before = snapCrud(hc) },
			AfterUpdate:  func(hc *CrudCtx) { afterRan = true; after = snapCrud(hc) },
		}

		plan := resourceGroupPlanState(ctx, b)
		resp := &resource.UpdateResponse{}
		b.Update(ctx, resource.UpdateRequest{Plan: tfsdk.Plan(plan)}, resp)

		if resp.Diagnostics.HasError() {
			t.Fatalf("Update returned errors: %v", resp.Diagnostics.Errors())
		}
		if !beforeRan {
			t.Fatal("BeforeUpdate hook was never invoked")
		}
		if !afterRan {
			t.Fatal("AfterUpdate hook was never invoked")
		}
		assertLiveness(t, "BeforeUpdate", before, liveness{planLive: true, stateLive: false, bodyLive: true, responseLive: false})
		assertLiveness(t, "AfterUpdate", after, liveness{planLive: true, stateLive: false, bodyLive: true, responseLive: true})
	})

	t.Run("read", func(t *testing.T) {
		desc := resourceGroupDescriptor()
		// read: single GET succeeds, so both BeforeRead and AfterRead fire (no
		// short-circuit -- the BeforeRead hook does NOT AddError here).
		b := &Base{desc: desc, provider: scriptClient(t, func(req *http.Request, _ int) (int, string) {
			return http.StatusOK, rgBody
		})}

		var beforeRan, afterRan bool
		var before, after liveness
		b.hooks = &Hooks{
			BeforeRead: func(hc *CrudCtx) { beforeRan = true; before = snapCrud(hc) },
			AfterRead:  func(hc *CrudCtx) { afterRan = true; after = snapCrud(hc) },
		}

		resp := &resource.ReadResponse{}
		b.Read(ctx, resource.ReadRequest{State: resourceGroupReadState(ctx, b, rgArmID)}, resp)

		if resp.Diagnostics.HasError() {
			t.Fatalf("Read returned errors: %v", resp.Diagnostics.Errors())
		}
		if !beforeRan {
			t.Fatal("BeforeRead hook was never invoked")
		}
		if !afterRan {
			t.Fatal("AfterRead hook was never invoked")
		}
		assertLiveness(t, "BeforeRead", before, liveness{planLive: false, stateLive: true, bodyLive: false, responseLive: false})
		assertLiveness(t, "AfterRead", after, liveness{planLive: false, stateLive: true, bodyLive: false, responseLive: true})
	})

	t.Run("delete", func(t *testing.T) {
		desc := resourceGroupDescriptor()
		// delete (non-singleton): DELETE -> 200.
		b := &Base{desc: desc, provider: scriptClient(t, func(req *http.Request, _ int) (int, string) {
			return http.StatusOK, `{}`
		})}

		var beforeRan bool
		var before liveness
		var afterRan bool
		var after liveness
		b.hooks = &Hooks{
			BeforeDelete: func(hc *CrudCtx) { beforeRan = true; before = snapCrud(hc) },
			AfterDelete:  func(hc *CrudCtx) { afterRan = true; after = snapCrud(hc) },
		}

		resp := &resource.DeleteResponse{}
		b.Delete(ctx, resource.DeleteRequest{State: resourceGroupReadState(ctx, b, rgArmID)}, resp)

		if resp.Diagnostics.HasError() {
			t.Fatalf("Delete returned errors: %v", resp.Diagnostics.Errors())
		}
		if !beforeRan {
			t.Fatal("BeforeDelete hook was never invoked")
		}
		if !afterRan {
			t.Fatal("AfterDelete hook was never invoked")
		}
		assertLiveness(t, "BeforeDelete", before, liveness{planLive: false, stateLive: true, bodyLive: false, responseLive: false})
		assertLiveness(t, "AfterDelete", after, liveness{planLive: false, stateLive: true, bodyLive: false, responseLive: false})
	})
}

// TestDeleteHookOverridesDefaultARMDelete pins the delete override seam for
// resources whose ARM type supports PUT/GET but has no management-plane DELETE.
// A custom Delete hook must run between BeforeDelete and AfterDelete, must receive
// the same delete-time CrudCtx liveness, and must replace -- not supplement -- the
// default ResourceClient.Delete call. The transport fails the test on any ARM
// request, so a regression to the default DELETE path reddens immediately.
func TestDeleteHookOverridesDefaultARMDelete(t *testing.T) {
	ctx := context.Background()
	desc := resourceGroupDescriptor()

	b := &Base{desc: desc, provider: scriptClient(t, func(req *http.Request, _ int) (int, string) {
		t.Fatalf("delete hook override issued an unexpected ARM request: %s %s", req.Method, req.URL.String())
		return http.StatusInternalServerError, `{"error":{"code":"UnexpectedARMCall"}}`
	})}

	var order []string
	var before, hookDelete, after liveness
	b.hooks = &Hooks{
		BeforeDelete: func(hc *CrudCtx) {
			order = append(order, "before")
			before = snapCrud(hc)
		},
		Delete: func(hc *CrudCtx) {
			order = append(order, "delete")
			hookDelete = snapCrud(hc)
		},
		AfterDelete: func(hc *CrudCtx) {
			order = append(order, "after")
			after = snapCrud(hc)
		},
	}

	resp := &resource.DeleteResponse{}
	b.Delete(ctx, resource.DeleteRequest{State: resourceGroupReadState(ctx, b, rgArmID)}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Delete returned errors: %v", resp.Diagnostics.Errors())
	}
	if got, want := strings.Join(order, ","), "before,delete,after"; got != want {
		t.Fatalf("delete hook order = %q, want %q", got, want)
	}
	want := liveness{planLive: false, stateLive: true, bodyLive: false, responseLive: false}
	assertLiveness(t, "BeforeDelete", before, want)
	assertLiveness(t, "Delete", hookDelete, want)
	assertLiveness(t, "AfterDelete", after, want)
}

// TestSingletonDeleteSkipsBeforeDelete pins the Singleton suppression rule: a
// Delete on a Singleton resource resets it via PUT (ARM has no DELETE for it) and
// that reset path bypasses both BeforeDelete and AfterDelete entirely. The reset
// PUT is the only ARM call, so neither delete hook has a place to fire; if a
// future change routes a Singleton destroy through either delete hook, this test
// reddens.
func TestSingletonDeleteSkipsBeforeDelete(t *testing.T) {
	ctx := context.Background()
	desc := resourceGroupDescriptor()

	// The only ARM call on a Singleton destroy is the reset PUT; answer it 200.
	b := &Base{desc: desc, provider: scriptClient(t, func(req *http.Request, _ int) (int, string) {
		return http.StatusOK, rgBody
	})}

	beforeRan := false
	afterRan := false
	b.hooks = &Hooks{
		Singleton:    &SingletonDefault{DefaultBody: map[string]interface{}{"properties": map[string]interface{}{}}},
		BeforeDelete: func(hc *CrudCtx) { beforeRan = true },
		AfterDelete:  func(hc *CrudCtx) { afterRan = true },
	}

	resp := &resource.DeleteResponse{}
	b.Delete(ctx, resource.DeleteRequest{State: resourceGroupReadState(ctx, b, rgArmID)}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Singleton Delete returned errors: %v", resp.Diagnostics.Errors())
	}
	if beforeRan {
		t.Fatal("BeforeDelete ran on the Singleton reset path; the contract says it must be bypassed")
	}
	if afterRan {
		t.Fatal("AfterDelete ran on the Singleton reset path; the contract says it must be bypassed")
	}
}
