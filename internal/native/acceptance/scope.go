package nativeacc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Azure/terraform-provider-azapi/internal/native/services"
	"github.com/hashicorp/terraform-exec/tfexec"
	gomega "github.com/onsi/gomega"
)

// Scope is the set of resources owned by one Ordered Ginkgo container. Resources in a
// scope are applied into the shared workspace and may reference resources from
// ancestor scopes by Terraform address. Teardown destroys only this scope's resources
// — removing their config and re-applying, which leaves ancestor (outer) resources
// running. This lets a nested container reuse the base resources of the containers
// that enclose it, and lets sibling containers each exercise a fresh resource on the
// same shared base.
//
// Nest scopes so that a resource's dependents live in DEEPER scopes than it: Ginkgo
// runs inner AfterAll (teardown) before outer, so dependents are destroyed first.
type Scope struct {
	ws        *Workspace
	owned     []string             // resource addresses (tfType.label) owned by this scope, in apply order
	ownedBy   map[string]*Resource // address -> owning resource, for collision detection
	tracked   []trackedResource    // applied resources, verified absent on Teardown
	ownedData []string             // data-source .tf file names owned by this scope

	// teardownRegistered records that this scope's Teardown was wired through the
	// workspace's teardown registrar when the scope was vended. The workspace safety
	// net asserts it: a vended scope left unregistered is one whose resources would
	// leak, so it is caught rather than silently skipped.
	teardownRegistered bool
}

// Scope returns a child scope and auto-registers its Teardown at the enclosing Ordered
// Ginkgo container, so the author who asks for a child scope gets its destroy wired
// without a paired AfterAll(child.Teardown) line. Because registration happens at this
// (the deeper) container, Ginkgo's inner-AfterAll-before-outer execution destroys a
// dependent nested in a deeper scope before the ancestor it depends on. Requesting a
// child scope outside an Ordered container (where teardown cannot be registered) fails
// loudly rather than leaking the scope's resources.
func (s *Scope) Scope() *Scope {
	child := &Scope{ws: s.ws}
	s.ws.registerScopeTeardown(child)
	return child
}

// Resource declares a resource-under-test owned by this scope. The ARM type and API
// version come from the generated descriptor, so the test targets the version the
// provider ships and the Azure existence check GETs at the matching API version.
func (s *Scope) Resource(tfType, label string) *Resource {
	return newResource(s, tfType, label)
}

// ResourceFor declares a resource-under-test in this scope from a config builder,
// taking the Terraform type and label from the config (see ResourceConfig). It pairs
// the handle with its config so neither the type nor the label is restated by the test.
func (s *Scope) ResourceFor(c ResourceConfig) *Resource {
	return s.Resource(c.ResourceType(), c.ResourceLabel())
}

// DataSource declares a data source in this scope's working directory: it writes the
// data source's HCL block to its own data_<type>.<label>.tf file so any resource in
// this or a deeper scope can reference it by address — e.g. the shared
// azapi_client_config every Key Vault scenario reads for its tenant/object id. Unlike
// a resource-under-test it is not applied on its own and not existence-checked
// (Terraform reads it during the next apply), and it is removed when this scope tears
// down. Declare a data source that dependents rely on in an ancestor scope that
// outlives them — typically the root scope, alongside the shared base resources.
func (s *Scope) DataSource(c DataSourceConfig) {
	name := dataSourceFileName(c.DataSourceType() + "." + c.Label())
	s.ws.writeFile(name, s.ws.render(c.Config()))
	for _, existing := range s.ownedData {
		if existing == name {
			return
		}
	}
	s.ownedData = append(s.ownedData, name)
}

// ApplyAll provisions several staged resources (Resource.Stage) in a SINGLE terraform
// apply, rather than one apply per resource. Terraform orders them from the
// cross-resource .id references in their configs, so a dependent chain can be created
// together — e.g. a resource group, its storage account and that account's blob
// service in one BeforeAll. One post-apply drift plan covers the whole set, then each
// staged resource's checks run. Each resource is recorded in its own scope (Stage
// carries the handle), so the receiver scope is just where the call reads naturally.
func (s *Scope) ApplyAll(staged ...Staged) {
	applyAll(s.ws, staged)
}

func newResource(s *Scope, tfType, label string) *Resource {
	d, ok := services.Registry[tfType]
	if !ok {
		panic(fmt.Sprintf("nativeacc: no generated descriptor for %q", tfType))
	}
	return &Resource{
		scope:      s,
		tfType:     tfType,
		label:      label,
		armType:    d.ARMType,
		apiVersion: d.APIVersion,
	}
}

// own records that this scope owns the file for the resource's address, so Teardown
// removes exactly the resources this scope created. Keyed by the full tfType.label
// address, not the label alone: distinct resource types may share a label
// (azapi_storage_account.test and azapi_resource_group.test are different resources
// that must not share a .tf file). Re-applying the SAME resource (create -> update) is
// idempotent; a DIFFERENT resource on an address this scope already owns is a label
// collision — a second instance of a type built without a distinct label — and panics
// loudly rather than silently clobbering the first resource's .tf file.
func (s *Scope) own(r *Resource) {
	addr := r.address()
	if prev, ok := s.ownedBy[addr]; ok {
		if prev != r {
			panic(fmt.Sprintf("nativeacc: two resources share the address %q in one scope; "+
				"give the second instance an explicit label, e.g. NewXxxCfg(parent, \"primary\")", addr))
		}
		return
	}
	if s.ownedBy == nil {
		s.ownedBy = map[string]*Resource{}
	}
	s.ownedBy[addr] = r
	s.owned = append(s.owned, addr)
}

// track records a resource for post-teardown existence verification, deduplicated by
// id so repeated Applies of the same resource are counted once.
func (s *Scope) track(tr trackedResource) {
	for _, t := range s.tracked {
		if t.id == tr.id {
			return
		}
	}
	s.tracked = append(s.tracked, tr)
}

// resourceFileName is the per-resource .tf file, keyed by the full tfType.label
// address so two resource types sharing a label get distinct files (the label alone
// is not unique across types). Neither component contains a dot, so the address is an
// injective, filesystem-safe file name.
func resourceFileName(address string) string { return "resource_" + address + ".tf" }

// dataSourceFileName is the per-data-source .tf file, keyed by the full type.label
// address and prefixed data_ so a data source and a resource sharing an address never
// collide on disk (mirrors resourceFileName).
func dataSourceFileName(address string) string { return "data_" + address + ".tf" }

// Teardown destroys the resources this scope owns and asserts they are gone from
// Azure, leaving ancestor resources running. It is auto-registered at the scope's
// Ordered container by Scope.Scope (the safety net and any bespoke ordering may still
// call it). It is a no-op when the workspace never started (a skipped run).
func (s *Scope) Teardown() {
	if s.ws == nil || !s.ws.started || s.ws.tf == nil {
		return
	}
	for _, address := range s.owned {
		_ = os.Remove(filepath.Join(s.ws.dir, resourceFileName(address)))
	}
	for _, name := range s.ownedData {
		_ = os.Remove(filepath.Join(s.ws.dir, name))
	}
	// s.ws.dumpTFConfigIfEnabled("before scoped teardown apply")
	if dumpTFConfigOnlyEnabled() {
		s.owned = nil
		s.ownedBy = nil
		s.tracked = nil
		s.ownedData = nil
		return
	}
	// Re-apply the remaining config: Terraform destroys the resources whose files we
	// just removed and no-ops everything still present (ancestor / sibling scopes).
	gomega.Expect(s.ws.tf.Apply(context.Background(), tfexec.Reattach(s.ws.reattach))).
		NotTo(gomega.HaveOccurred(), "scoped teardown apply")

	for _, tr := range s.tracked {
		if tr.id == "" {
			continue
		}
		ok, err := azureExists(s.ws.client, tr.armType, tr.apiVersion, tr.id)
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "post-teardown GET %s", tr.id)
		gomega.Expect(ok).To(gomega.BeFalse(), "%s still exists after scope teardown", tr.id)
	}
	s.owned = nil
	s.ownedBy = nil
	s.tracked = nil
	s.ownedData = nil
}
