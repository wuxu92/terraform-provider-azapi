package nativeacc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
// Resource handle, so the literal type and label are spelled exactly once — set from
// the resource's generated Descriptor.Name in the builder's NewXxxCfg constructor —
// not restated here.
type ResourceConfig interface {
	ResourceType() string     // Terraform type, e.g. "azapi_storage_account"
	ResourceLabel() string    // Terraform state label
	IDRef() string            // Terraform reference to the resource's id attribute, e.g. "azapi_resource_group.rg.id".
	RefOf(path string) string // Terraform reference to an arbitrary attribute, e.g. "azapi_resource_group.rg.location" for RefOf("location"). It derives from the same type.label as the acceptance Resource handle's IDRef
	Config() string           // HCL config block for the resource, rendered by the builder
}

type DataSourceConfig interface {
	DataSourceType() string   // Terraform type, e.g. "azapi_storage_account"
	Label() string            // Terraform state label
	IDRef() string            // Terraform reference to the resource's id attribute, e.g. "azapi_resource_group.rg.id".
	RefOf(path string) string // Terraform reference to an arbitrary attribute, e.g. "azapi_resource_group.rg.location" for RefOf("location"). It derives from the same type.label as the acceptance Resource handle's IDRef
	Config() string           // HCL config block for the resource, rendered by the builder
}

// Apply writes the resource's config block, applies the workspace, and then re-plans
// to assert the apply left no drift: a refresh-backed plan must be empty, proving the
// config round-trips through the provider (Create/Update -> Read -> plan is stable).
// This generic drift check catches the bulk of round-trip regressions, so the checks
// argument is reserved for facts the plan can't prove — Azure-side existence (Exists)
// and computed defaults (a value the provider, not the config, supplies). The resource
// is created on first Apply and updated in place on subsequent Applies (same address).
// config is either a literal HCL string or a value exposing Config() string (e.g. a
// generated config builder), so a spec can pass the builder instance directly.
func (r *Resource) Apply(config any, checks ...Check) *Resource {
	applyAll(r.scope.ws, []Staged{r.Stage(config, checks...)})
	return r
}

// Staged pairs a resource with the config and checks to apply for it. It is produced
// by Resource.Stage and consumed by ApplyAll, which writes every staged config and
// then runs a SINGLE terraform apply for the whole set.
type Staged struct {
	r      *Resource
	config any
	checks []Check
}

// Stage prepares this resource's config and checks for a batched apply via
// Scope.ApplyAll / Workspace.ApplyAll, instead of Apply's one-terraform-apply-per-
// resource. Use it to provision several resources — typically a dependent chain — in
// one apply, e.g. a whole base in a single BeforeAll. The config is resolved exactly
// as in Apply (a literal HCL string or a value exposing Config() string).
func (r *Resource) Stage(config any, checks ...Check) Staged {
	return Staged{r: r, config: config, checks: checks}
}

// applyAll writes every staged resource's config, runs ONE terraform apply to create
// or update them together (Terraform orders them from the cross-resource .id
// references in their configs), asserts the post-apply plan is empty for the whole
// working directory, then runs each staged resource's checks. Apply is the single-
// resource case; Scope.ApplyAll / Workspace.ApplyAll expose the batch to specs.
func applyAll(w *Workspace, staged []Staged) {
	gomega.Expect(staged).NotTo(gomega.BeEmpty(), "ApplyAll requires at least one staged resource")
	ctx := context.Background()
	addrs := make([]string, len(staged))
	for i, s := range staged {
		gomega.Expect(s.r.scope.ws).To(gomega.BeIdenticalTo(w),
			"staged resource %s belongs to a different workspace", s.r.address())
		s.r.scope.own(s.r)
		s.r.write(configHCL(s.config))
		addrs[i] = s.r.address()
	}
	label := strings.Join(addrs, ", ")
	gomega.Expect(w.tf.Apply(ctx, tfexec.Reattach(w.reattach))).
		NotTo(gomega.HaveOccurred(), "terraform apply (%s)", label)
	hasChanges, err := w.tf.Plan(ctx, tfexec.Reattach(w.reattach))
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "plan after apply (%s)", label)
	gomega.Expect(hasChanges).To(gomega.BeFalse(), "plan after apply shows drift for %s", label)
	for _, s := range staged {
		s.r.verify(s.checks, true)
	}
}

// configHCL resolves Apply's config argument: a literal HCL string, or any value that
// renders its own block via Config() string — e.g. a generated config builder, letting
// a spec pass the builder instance directly instead of builder.Basic().
func configHCL(config any) string {
	switch c := config.(type) {
	case string:
		return c
	case interface{ Config() string }:
		return c.Config()
	default:
		panic(fmt.Sprintf("nativeacc: Apply config must be a string or implement Config() string, got %T", config))
	}
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
	defer func() { _ = os.Remove(filepath.Join(r.scope.ws.dir, resourceFileName(r.address()))) }()
	_, err := r.scope.ws.tf.Plan(context.Background(), tfexec.Reattach(r.scope.ws.reattach))
	gomega.Expect(err).To(gomega.HaveOccurred(), "expected plan to fail for %s", r.address())
	gomega.Expect(err.Error()).To(gomega.MatchRegexp(errRegex))
}

func (r *Resource) write(config string) {
	r.scope.ws.writeFile(resourceFileName(r.address()), r.scope.ws.render(config))
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
