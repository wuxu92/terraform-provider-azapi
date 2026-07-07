package nativeacc

import (
	"os"
	"path/filepath"
	"testing"

	gomega "github.com/onsi/gomega"
)

// fakeDataSource is a self-contained DataSourceConfig double: it names a Terraform
// type/label and renders a fixed, token-free data block. Config() deliberately has no
// template tokens so render leaves it byte-for-byte unchanged, letting the tests assert
// the exact on-disk content. IDRef/RefOf are unused by DataSource but derived so the
// double satisfies the full interface without pulling in the services/config package.
type fakeDataSource struct {
	tfType string
	label  string
}

func (f fakeDataSource) DataSourceType() string { return f.tfType }
func (f fakeDataSource) Label() string          { return f.label }
func (f fakeDataSource) Config() string {
	return "data \"" + f.tfType + "\" \"" + f.label + "\" {}"
}
func (f fakeDataSource) IDRef() string { return "data." + f.tfType + "." + f.label + ".id" }
func (f fakeDataSource) RefOf(path string) string {
	return "data." + f.tfType + "." + f.label + "." + path
}

// TestDataSourceFileNameKeyedByAddress mirrors TestResourceFileNameKeyedByAddress for
// the data-source file namer: the file name is keyed by the full type.label address and
// prefixed data_, so a data source and a resource that happen to share an address never
// collide on disk (the data_ vs resource_ prefixes keep them distinct).
func TestDataSourceFileNameKeyedByAddress(t *testing.T) {
	const addr = "azapi_client_config.current"
	ds := dataSourceFileName(addr)
	if want := "data_azapi_client_config.current.tf"; ds != want {
		t.Errorf("file name = %q, want %q", ds, want)
	}
	if rs := resourceFileName(addr); ds == rs {
		t.Fatalf("data source and resource sharing address %q collide on file name: %q", addr, ds)
	}
}

// TestScopeDataSourceWritesAndOwns proves DataSource does the two things Teardown then
// depends on: it renders the block to the address-keyed data_ file on disk, and it
// records that file name in the scope's ownedData. A repeat declaration of the SAME
// address is a no-op — the file is rewritten but ownedData does not grow — so Teardown
// removes the file exactly once.
func TestScopeDataSourceWritesAndOwns(t *testing.T) {
	gomega.RegisterTestingT(t)
	ws := NewWorkspace()
	ws.dir = t.TempDir()

	fake := fakeDataSource{tfType: "azapi_client_config", label: "current"}
	ws.DataSource(fake)

	name := "data_azapi_client_config.current.tf"
	content, err := os.ReadFile(filepath.Join(ws.dir, name))
	if err != nil {
		t.Fatalf("reading written data-source file: %v", err)
	}
	if got, want := string(content), fake.Config(); got != want {
		t.Errorf("file contents = %q, want %q (render must leave a token-free block unchanged)", got, want)
	}

	if got := ws.root.ownedData; len(got) != 1 || got[0] != name {
		t.Fatalf("ownedData = %v, want exactly [%q]", got, name)
	}

	ws.DataSource(fake) // same address re-declared — must be a no-op, not a second entry
	if got := ws.root.ownedData; len(got) != 1 {
		t.Fatalf("re-declaring the same data source should dedup; ownedData = %v", got)
	}
}

// TestScopeDataSourceDistinctAddressesOwnedSeparately guards that dedup is keyed by
// address, not a blanket suppression: two DIFFERENT data-source addresses declared in
// one scope are tracked as two distinct ownedData entries, so Teardown removes both.
func TestScopeDataSourceDistinctAddressesOwnedSeparately(t *testing.T) {
	gomega.RegisterTestingT(t)
	ws := NewWorkspace()
	ws.dir = t.TempDir()

	ws.DataSource(fakeDataSource{tfType: "azapi_client_config", label: "current"})
	ws.DataSource(fakeDataSource{tfType: "azapi_resources", label: "existing"})

	got := ws.root.ownedData
	if len(got) != 2 {
		t.Fatalf("two distinct data-source addresses should yield 2 ownedData entries, got %d: %v", len(got), got)
	}
	want := map[string]bool{
		"data_azapi_client_config.current.tf": true,
		"data_azapi_resources.existing.tf":    true,
	}
	for _, name := range got {
		if !want[name] {
			t.Errorf("unexpected ownedData entry %q", name)
		}
	}
}
