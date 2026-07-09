package nativeacc

import (
	"context"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	gomega "github.com/onsi/gomega"
)

// DataSourceUnderTest is a data source read back and asserted after an apply. Unlike
// Scope.DataSource (which declares a data source that OTHER resources reference and is
// never asserted on its own), this handle owns a data block that reads an existing
// resource and lets a spec Check the data source's own flattened state — the read path
// of a native data source end to end (schema, id compose, GET, mapper flatten).
//
// The data block is written on construction, so a subsequent Apply in the same scope
// (typically the resource-under-test's Apply) reads it during that apply. Check then
// reads the data source's state at address data.<type>.<label>.
type DataSourceUnderTest struct {
	scope  *Scope
	tfType string
	label  string
}

// DataSourceUnderTest declares a data source-under-test in this scope: it writes the
// data block to its data_<address>.tf file (so the next apply reads it) and returns a
// handle whose Check asserts against the data source's read-back state. The block is
// removed when the scope tears down, like Scope.DataSource.
func (s *Scope) DataSourceUnderTest(c DataSourceConfig) *DataSourceUnderTest {
	name := dataSourceFileName(c.DataSourceType() + "." + c.Label())
	s.ws.writeFile(name, s.ws.render(c.Config()))
	for _, existing := range s.ownedData {
		if existing == name {
			return &DataSourceUnderTest{scope: s, tfType: c.DataSourceType(), label: c.Label()}
		}
	}
	s.ownedData = append(s.ownedData, name)
	return &DataSourceUnderTest{scope: s, tfType: c.DataSourceType(), label: c.Label()}
}

// DataSourceUnderTest declares a data source-under-test in the workspace root scope,
// mirroring Workspace.ResourceFor.
func (w *Workspace) DataSourceUnderTest(c DataSourceConfig) *DataSourceUnderTest {
	return w.root.DataSourceUnderTest(c)
}

func (d *DataSourceUnderTest) address() string { return "data." + d.tfType + "." + d.label }

// Check reads the data source's state (populated by a prior apply that read the data
// block) and runs the checks against its attributes. Existence-in-Azure checks do not
// apply to a data source, so a Check that calls exists panics — assert Azure-side facts
// on the resource-under-test instead.
func (d *DataSourceUnderTest) Check(checks ...Check) *DataSourceUnderTest {
	if dumpTFConfigOnlyEnabled() {
		return d
	}
	res := d.stateResource()
	gomega.Expect(res).NotTo(gomega.BeNil(), "data source %s not found in state", d.address())
	id, _ := res.AttributeValues["id"].(string)
	c := &checkCtx{
		attrs: res.AttributeValues,
		id:    id,
		exists: func() (bool, error) {
			panic("nativeacc: Exists() is not applicable to a data source under test")
		},
	}
	for _, ch := range checks {
		ch(c)
	}
	return d
}

func (d *DataSourceUnderTest) stateResource() *tfjson.StateResource {
	st, err := d.scope.ws.tf.Show(context.Background(), tfexec.Reattach(d.scope.ws.reattach))
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "terraform show")
	if st == nil || st.Values == nil || st.Values.RootModule == nil {
		return nil
	}
	for _, res := range st.Values.RootModule.Resources {
		if res.Address == d.address() {
			return res
		}
	}
	return nil
}
