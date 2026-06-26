package nativeacc

import (
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/native/generated"
)

// TestResourceFileNameKeyedByAddress guards against resource .tf files colliding when
// two resource types share a Terraform label: the file name is keyed by the full
// tfType.label address, not the label alone, so it stays unique across types.
func TestResourceFileNameKeyedByAddress(t *testing.T) {
	sa := resourceFileName("azapi_storage_account.test")
	rg := resourceFileName("azapi_resource_group.test")
	if sa == rg {
		t.Fatalf("same-label/different-type resources share a file name: %q", sa)
	}
	if want := "resource_azapi_storage_account.test.tf"; sa != want {
		t.Errorf("file name = %q, want %q", sa, want)
	}
}

// TestScopeOwnKeyedByAddress guards that a scope tracks owned files by the owning
// resource's address: distinct addresses are kept (even with a shared label across
// types, the case that motivated the file-keying fix), while re-applying the SAME
// resource records it once so Teardown removes each file exactly once.
func TestScopeOwnKeyedByAddress(t *testing.T) {
	var s Scope
	sa := &Resource{scope: &s, tfType: "azapi_storage_account", label: "test"}
	rg := &Resource{scope: &s, tfType: "azapi_resource_group", label: "test"}
	s.own(sa)
	s.own(rg) // same label, different type — must stay distinct
	s.own(sa) // same resource re-applied — deduped
	if len(s.owned) != 2 {
		t.Fatalf("own should retain 2 distinct addresses, got %d: %v", len(s.owned), s.owned)
	}
}

// TestScopeOwnRejectsLabelCollision guards Option C's safety net: two DISTINCT
// resources that land on the same tfType.label address in one scope are a label
// collision (a second instance built without a distinct label) and must panic loudly
// rather than silently clobber the first resource's .tf file.
func TestScopeOwnRejectsLabelCollision(t *testing.T) {
	var s Scope
	first := &Resource{scope: &s, tfType: "azapi_storage_account", label: "test"}
	dup := &Resource{scope: &s, tfType: "azapi_storage_account", label: "test"}
	s.own(first)
	defer func() {
		if recover() == nil {
			t.Fatal("own should panic on a duplicate address from a distinct resource")
		}
	}()
	s.own(dup)
}

// TestConfigHCL guards Apply's config resolver: a literal HCL string passes through, a
// value exposing Config() string is rendered via that method (so a spec can pass a
// builder instance directly), and anything else panics loudly.
func TestConfigHCL(t *testing.T) {
	const literal = "resource azapi_resource_group rg { name = \"x\" }"
	if got := configHCL(literal); got != literal {
		t.Errorf("literal string: got %q, want %q", got, literal)
	}
	cfg := generated.NewResourceConfigBase("azapi_resource_group", "rg")
	if got, want := configHCL(cfg), "resource azapi_resource_group rg {}"; got != want {
		t.Errorf("Config() builder: got %q, want %q", got, want)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("configHCL should panic on an unsupported config type")
		}
	}()
	configHCL(42)
}
