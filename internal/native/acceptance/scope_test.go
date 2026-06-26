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

// TestScopeOwnKeyedByAddress guards that a scope tracks owned files by address:
// distinct addresses are kept (even with a shared label, the case that motivated the
// fix), while a repeated address is recorded once so Teardown removes each file
// exactly once.
func TestScopeOwnKeyedByAddress(t *testing.T) {
	var s Scope
	s.own("azapi_storage_account.test")
	s.own("azapi_resource_group.test")  // same label, different type — must stay distinct
	s.own("azapi_storage_account.test") // repeat of the first — deduped
	if len(s.owned) != 2 {
		t.Fatalf("own should retain 2 distinct addresses, got %d: %v", len(s.owned), s.owned)
	}
}
