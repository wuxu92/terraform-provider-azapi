package nativeacc

import (
	"bytes"
	"context"
	"encoding/json"
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

// Apply writes the resource's config block, applies the workspace, and then re-plans
// to assert the apply left no drift: a refresh-backed plan must be empty, proving the
// config round-trips through the provider (Create/Update -> Read -> plan is stable).
// This generic drift check catches the bulk of round-trip regressions, so the checks
// argument is reserved for facts the plan can't prove — Azure-side existence (Exists)
// and computed defaults (a value the provider, not the config, supplies). The resource
// is created on first Apply and updated in place on subsequent Applies (same address).
// config is a Configure scenario value. Use StringConfigure for one-off literal HCL.
func (r *Resource) Apply(config Configure, checks ...Check) *Resource {
	applyAll(r.scope.ws, []Staged{r.Stage(config, checks...)})
	return r
}

// Staged pairs a resource with the config and checks to apply for it. It is produced
// by Resource.Stage and consumed by ApplyAll, which writes every staged config and
// then runs a SINGLE terraform apply for the whole set.
type Staged struct {
	r      *Resource
	config Configure
	checks []Check
}

// Stage prepares this resource's config and checks for a batched apply via
// Scope.ApplyAll / Workspace.ApplyAll, instead of Apply's one-terraform-apply-per-
// resource. Use it to provision several resources — typically a dependent chain — in
// one apply, e.g. a whole base in a single BeforeAll. The config is resolved exactly
// as in Apply (a Configure scenario value; use StringConfigure for one-off literal HCL).
func (r *Resource) Stage(config Configure, checks ...Check) Staged {
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
		s.r.write(s.config.Config())
		addrs[i] = s.r.address()
	}
	label := strings.Join(addrs, ", ")
	w.dumpTFConfigIfEnabled("before terraform apply (" + label + ")")
	if dumpTFConfigOnlyEnabled() {
		return
	}
	gomega.Expect(w.tf.Apply(ctx, tfexec.Reattach(w.reattach))).
		NotTo(gomega.HaveOccurred(), "terraform apply (%s)", label)
	expectNoPlanDrift(ctx, w, label, "plan after apply")
	for _, s := range staged {
		s.r.verify(s.checks, true)
	}
}

// Check runs checks against the resource's current state / Azure without re-applying.
// Use it in an It that asserts a facet of a resource provisioned in the container's
// BeforeAll.
func (r *Resource) Check(checks ...Check) *Resource {
	if dumpTFConfigOnlyEnabled() {
		return r
	}
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
	if dumpTFConfigOnlyEnabled() {
		return r
	}
	ctx := context.Background()
	id := r.currentID()
	gomega.Expect(id).NotTo(gomega.BeEmpty(), "resource %s has no id to import", r.address())

	gomega.Expect(r.scope.ws.tf.StateRm(ctx, r.address())).
		NotTo(gomega.HaveOccurred(), "state rm %s", r.address())
	gomega.Expect(r.scope.ws.tf.Import(ctx, r.address(), id, tfexec.Reattach(r.scope.ws.reattach))).
		NotTo(gomega.HaveOccurred(), "import %s", r.address())

	expectNoPlanDrift(ctx, r.scope.ws, r.address(), "plan after import")
	return r
}

// ApplyExpectError writes config and expects terraform plan to fail with an error
// matching errRegex. This targets provider-side (schema) validation, which fails
// during plan and never reaches Azure; the config file is removed afterward so it
// does not affect later scenarios. Use a label distinct from any resource-under-test
// that was (or will be) applied.
func (r *Resource) ApplyExpectError(config Configure, errRegex string) {
	r.write(config.Config())
	// Remove the config even if an assertion below fails (gomega panics on failure),
	// so a regressed negative case cannot leave an invalid .tf that poisons teardown.
	defer func() { _ = os.Remove(filepath.Join(r.scope.ws.dir, resourceFileName(r.address()))) }()
	r.scope.ws.dumpTFConfigIfEnabled("before terraform plan expecting error (" + r.address() + ")")
	if dumpTFConfigOnlyEnabled() {
		return
	}
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

func expectNoPlanDrift(ctx context.Context, w *Workspace, label, operation string) {
	var planJSON bytes.Buffer
	w.tf.SetStdout(&planJSON)
	// defer w.tf.SetStdout(os.Stdout)
	hasChanges, err := w.tf.Plan(ctx, tfexec.Reattach(w.reattach))
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "%s (%s)", operation, label)
	gomega.Expect(hasChanges).To(gomega.BeFalse(), "%s shows drift for %s\n%s", operation, label, formatPlanJSONDrift(planJSON.String()))
}

func formatPlanJSONDrift(output string) string {
	lines := strings.Split(output, "\n")
	summary := make([]string, 0)
	diagnostics := make([]string, 0)
	changes := make([]string, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var msg map[string]interface{}
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		switch msg["type"] {
		case "change_summary":
			if c, ok := msg["changes"].(map[string]interface{}); ok {
				summary = append(summary, fmt.Sprintf("change summary: add=%v change=%v remove=%v operation=%v", c["add"], c["change"], c["remove"], msg["operation"]))
			}
		case "planned_change", "resource_drift":
			if c, ok := msg["change"].(map[string]interface{}); ok {
				addr := "<unknown>"
				if r, ok := c["resource"].(map[string]interface{}); ok {
					if v, ok := r["addr"].(string); ok && v != "" {
						addr = v
					}
				}
				action := c["action"]
				if actions, ok := c["actions"]; ok {
					action = actions
				}
				changes = append(changes, fmt.Sprintf("%s: %s action=%v reason=%v", msg["type"], addr, action, c["reason"]))
			}
		case "diagnostic":
			if d, ok := msg["diagnostic"].(map[string]interface{}); ok {
				diagnostics = append(diagnostics, fmt.Sprintf("diagnostic: severity=%v summary=%v detail=%v", d["severity"], d["summary"], d["detail"]))
			}
		}
	}

	parts := make([]string, 0, 4)
	parts = append(parts, summary...)
	parts = append(parts, changes...)
	parts = append(parts, diagnostics...)
	if len(parts) == 0 {
		return "Terraform plan JSON contained changes but no parsed drift details. Raw output:\n" + output
	}
	parts = append(parts, "raw terraform plan JSON:", output)
	return strings.Join(parts, "\n")
}

func (r *Resource) address() string { return r.tfType + "." + r.label }

// Address is the resource's Terraform address (tfType.label), e.g.
// "azapi_resource_group.rg".
func (r *Resource) Address() string { return r.address() }

// IDRef is the Terraform reference to this resource's id attribute, e.g.
// "azapi_resource_group.rg.id". Pass it to a dependent resource's config builder so
// the dependent injects this resource as its parent by address.
func (r *Resource) IDRef() string { return r.address() + ".id" }
