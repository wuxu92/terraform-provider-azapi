package nativeacc

import (
	"context"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	gomega "github.com/onsi/gomega"
)

// Resource is a resource-under-test handle owned by a Scope. Each It drives one
// Resource: Apply a full resource-block config (which also asserts no post-apply
// drift), Check the current state, ImportVerify it, or ApplyExpectError. The resource
// is destroyed when its owning scope tears down.
type Resource struct {
	scope *Scope

	tfType     string
	label      string
	armType    string
	apiVersion string
}

// ResourceConfig is implemented by a generated config builder via the embedded
// generated.ResourceConfigBase: it names the Terraform type and state label of the
// resource it configures. A scope's ResourceFor takes one and vends the matching
// Resource handle, so the literal type and label are spelled exactly once — set from a
// generated.Type* constant in the builder's NewXxx constructor — not restated here.
type ResourceConfig interface {
	ResourceType() string  // Terraform type, e.g. "azapi_storage_account"
	ResourceLabel() string // Terraform state label
}

// Apply writes the resource's config block, applies the workspace, and then re-plans
// to assert the apply left no drift: a refresh-backed plan must be empty, proving the
// config round-trips through the provider (Create/Update -> Read -> plan is stable).
// This generic drift check catches the bulk of round-trip regressions, so the checks
// argument is reserved for facts the plan can't prove — Azure-side existence (Exists)
// and computed defaults (a value the provider, not the config, supplies). The resource
// is created on first Apply and updated in place on subsequent Applies (same label).
func (r *Resource) Apply(config string, checks ...Check) *Resource {
	ctx := context.Background()
	r.scope.own(r.label)
	r.write(config)
	gomega.Expect(r.scope.ws.tf.Apply(ctx, tfexec.Reattach(r.scope.ws.reattach))).
		NotTo(gomega.HaveOccurred(), "terraform apply (%s)", r.address())
	hasChanges, err := r.scope.ws.tf.Plan(ctx, tfexec.Reattach(r.scope.ws.reattach))
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "plan after apply (%s)", r.address())
	gomega.Expect(hasChanges).To(gomega.BeFalse(), "plan after apply shows drift for %s", r.address())
	r.verify(checks, true)
	return r
}

// Check runs checks against the resource's current state / Azure without re-applying.
// Use it in an It that asserts a facet of a resource provisioned in the container's
// BeforeAll.
func (r *Resource) Check(checks ...Check) *Resource {
	r.verify(checks, false)
	return r
}

// ImportVerify removes the resource from state, re-imports it by ID, and asserts the
// subsequent plan is empty — i.e. the read/import path reproduces the configured state
// with no drift. Call it in the same It/BeforeAll immediately after the Apply it
// verifies (like azurerm's data.ImportStep, which is the step right after the apply
// step) — never in a separate It, which would silently depend on another spec having
// applied the resource first.
func (r *Resource) ImportVerify() *Resource {
	ctx := context.Background()
	id := r.currentID()
	gomega.Expect(id).NotTo(gomega.BeEmpty(), "resource %s has no id to import", r.address())

	gomega.Expect(r.scope.ws.tf.StateRm(ctx, r.address())).
		NotTo(gomega.HaveOccurred(), "state rm %s", r.address())
	gomega.Expect(r.scope.ws.tf.Import(ctx, r.address(), id, tfexec.Reattach(r.scope.ws.reattach))).
		NotTo(gomega.HaveOccurred(), "import %s", r.address())

	hasChanges, err := r.scope.ws.tf.Plan(ctx, tfexec.Reattach(r.scope.ws.reattach))
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "plan after import")
	gomega.Expect(hasChanges).To(gomega.BeFalse(), "plan after import shows drift for %s", r.address())
	return r
}

// ApplyExpectError writes config and expects terraform plan to fail with an error
// matching errRegex. This targets provider-side (schema) validation, which fails
// during plan and never reaches Azure; the config file is removed afterward so it
// does not affect later scenarios. Use a label distinct from any resource-under-test
// that was (or will be) applied.
func (r *Resource) ApplyExpectError(config, errRegex string) {
	r.write(config)
	// Remove the config even if an assertion below fails (gomega panics on failure),
	// so a regressed negative case cannot leave an invalid .tf that poisons teardown.
	defer func() { _ = os.Remove(filepath.Join(r.scope.ws.dir, resourceFileName(r.label))) }()
	_, err := r.scope.ws.tf.Plan(context.Background(), tfexec.Reattach(r.scope.ws.reattach))
	gomega.Expect(err).To(gomega.HaveOccurred(), "expected plan to fail for %s", r.address())
	gomega.Expect(err.Error()).To(gomega.MatchRegexp(errRegex))
}

func (r *Resource) write(config string) {
	r.scope.ws.writeFile(resourceFileName(r.label), r.scope.ws.render(config))
}

// verify reads the resource from state, optionally records it for post-teardown
// existence verification, and runs the checks.
func (r *Resource) verify(checks []Check, track bool) {
	res := r.stateResource()
	gomega.Expect(res).NotTo(gomega.BeNil(), "resource %s not found in state", r.address())

	id, _ := res.AttributeValues["id"].(string)
	if track {
		r.scope.track(trackedResource{armType: r.armType, apiVersion: r.apiVersion, id: id})
	}

	c := &checkCtx{
		attrs: res.AttributeValues,
		id:    id,
		exists: func() (bool, error) {
			return azureExists(r.scope.ws.client, r.armType, r.apiVersion, id)
		},
	}
	for _, ch := range checks {
		ch(c)
	}
}

func (r *Resource) currentID() string {
	res := r.stateResource()
	if res == nil {
		return ""
	}
	id, _ := res.AttributeValues["id"].(string)
	return id
}

func (r *Resource) stateResource() *tfjson.StateResource {
	st, err := r.scope.ws.tf.Show(context.Background(), tfexec.Reattach(r.scope.ws.reattach))
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "terraform show")
	if st == nil || st.Values == nil || st.Values.RootModule == nil {
		return nil
	}
	for _, res := range st.Values.RootModule.Resources {
		if res.Address == r.address() {
			return res
		}
	}
	return nil
}

func (r *Resource) address() string { return r.tfType + "." + r.label }

// Address is the resource's Terraform address (tfType.label), e.g.
// "azapi_resource_group.rg".
func (r *Resource) Address() string { return r.address() }

// IDRef is the Terraform reference to this resource's id attribute, e.g.
// "azapi_resource_group.rg.id". Pass it to a dependent resource's config builder so
// the dependent injects this resource as its parent by address.
func (r *Resource) IDRef() string { return r.address() + ".id" }
