package nativeacc

import "testing"

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

// TestStringConfigure adapts raw HCL into the Configure interface without keeping a
// separate config resolver in Resource.Apply.
func TestStringConfigure(t *testing.T) {
	const hcl = "resource azapi_resource_group rg {}"
	var cfg Configure = StringConfigure(hcl)
	if got := cfg.Config(); got != hcl {
		t.Errorf("StringConfigure.Config() = %q, want %q", got, hcl)
	}
}

// TestStageBundlesConfigAndChecks guards Resource.Stage, the batched-apply building
// block: it must bundle the handle, config and checks verbatim so ApplyAll writes the
// right config and runs the right checks per resource. The apply itself drives
// Terraform and is exercised by the Azure-gated acceptance suite.
func TestStageBundlesConfigAndChecks(t *testing.T) {
	var s Scope
	r := &Resource{scope: &s, tfType: "azapi_resource_group", label: "rg"}
	chk := func(*checkCtx) {}
	cfg := StringConfigure("cfg-hcl")
	st := r.Stage(cfg, chk, chk)
	if st.r != r {
		t.Errorf("Stage dropped the resource handle")
	}
	if st.config != cfg {
		t.Errorf("Stage config = %v, want %q", st.config, cfg)
	}
	if len(st.checks) != 2 {
		t.Errorf("Stage checks = %d, want 2", len(st.checks))
	}
}
