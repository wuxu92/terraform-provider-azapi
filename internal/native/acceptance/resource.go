package nativeacc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Azure/terraform-provider-azapi/internal/native/generated"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	gomega "github.com/onsi/gomega"
)

// Resource is a resource-under-test handle within a Workspace. Each It drives one
// Resource: Apply a full resource-block config and assert, ImportVerify it, or
// ApplyExpectError. The resource lives in the shared workspace and is torn down
// with it.
type Resource struct {
	w *Workspace

	tfType     string
	label      string
	armType    string
	apiVersion string
}

// Resource declares the resource-under-test (Terraform type + state label) in the
// workspace. The ARM type and API version are read from the resource's generated
// descriptor, so the test always targets the version the provider ships and the
// Azure existence check GETs at the matching API version. The config passed to
// Apply is a complete `resource "<type>" "<label>" { ... }` block.
func (w *Workspace) Resource(tfType, label string) *Resource {
	d, ok := generated.Registry[tfType]
	if !ok {
		panic(fmt.Sprintf("nativeacc: no generated descriptor for %q", tfType))
	}
	return &Resource{
		w:          w,
		tfType:     tfType,
		label:      label,
		armType:    d.ARMType,
		apiVersion: d.APIVersion,
	}
}

// Apply writes the resource's config block, applies the workspace, and runs the
// checks against the resulting state / Azure. The resource is created on first
// Apply and updated in place on subsequent Applies (same label); pass an updated
// config to drive an update.
func (r *Resource) Apply(config string, checks ...Check) {
	r.write(config)
	gomega.Expect(r.w.tf.Apply(context.Background(), tfexec.Reattach(r.w.reattach))).
		NotTo(gomega.HaveOccurred(), "terraform apply (%s)", r.address())
	r.runChecks(checks)
}

// ImportVerify removes the resource from state, re-imports it by ID, and asserts
// the subsequent plan is empty — i.e. the read/import path reproduces the
// configured state with no drift.
func (r *Resource) ImportVerify() {
	ctx := context.Background()
	id := r.currentID()
	gomega.Expect(id).NotTo(gomega.BeEmpty(), "resource %s has no id to import", r.address())

	gomega.Expect(r.w.tf.StateRm(ctx, r.address())).
		NotTo(gomega.HaveOccurred(), "state rm %s", r.address())
	gomega.Expect(r.w.tf.Import(ctx, r.address(), id, tfexec.Reattach(r.w.reattach))).
		NotTo(gomega.HaveOccurred(), "import %s", r.address())

	hasChanges, err := r.w.tf.Plan(ctx, tfexec.Reattach(r.w.reattach))
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "plan after import")
	gomega.Expect(hasChanges).To(gomega.BeFalse(), "plan after import shows drift for %s", r.address())
}

// ApplyExpectError writes config and expects terraform plan to fail with an error
// matching errRegex. This targets provider-side (schema) validation, which fails
// during plan and never reaches Azure; the config file is removed afterward so it
// does not affect later scenarios. Use a label distinct from any resource-under-test
// that was (or will be) applied.
func (r *Resource) ApplyExpectError(config, errRegex string) {
	r.write(config)
	_, err := r.w.tf.Plan(context.Background(), tfexec.Reattach(r.w.reattach))
	gomega.Expect(err).To(gomega.HaveOccurred(), "expected plan to fail for %s", r.address())
	gomega.Expect(err.Error()).To(gomega.MatchRegexp(errRegex))
	_ = os.Remove(filepath.Join(r.w.dir, r.fileName()))
}

func (r *Resource) write(config string) {
	r.w.writeFile(r.fileName(), r.w.render(config))
}

func (r *Resource) runChecks(checks []Check) {
	res := r.stateResource()
	gomega.Expect(res).NotTo(gomega.BeNil(), "resource %s not found in state", r.address())

	id, _ := res.AttributeValues["id"].(string)
	r.w.track(r.armType, r.apiVersion, id)

	c := &checkCtx{
		attrs: res.AttributeValues,
		id:    id,
		exists: func() (bool, error) {
			return azureExists(r.w.client, r.armType, r.apiVersion, id)
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
	st, err := r.w.tf.Show(context.Background(), tfexec.Reattach(r.w.reattach))
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

func (r *Resource) fileName() string { return "resource_" + r.label + ".tf" }
func (r *Resource) address() string  { return r.tfType + "." + r.label }
