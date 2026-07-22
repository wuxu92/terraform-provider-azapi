package storage

import "testing"

// TestStorageForceNewPredicates locks the value-conditional replacement rules the
// storage account ModifyPlan hook enforces (inlined from the former azwise
// CheckForceNew override): SKU zonal/non-zonal boundary crossing, account_kind
// migration (only Storage -> StorageV2 in place), and large-file-share disablement.
func TestStorageForceNewPredicates(t *testing.T) {
	t.Run("skuZoneMigration", func(t *testing.T) {
		cases := []struct {
			old, new string
			want     bool
		}{
			{"Standard_LRS", "Standard_LRS", false},  // unchanged
			{"Standard_LRS", "Standard_ZRS", true},   // non-zonal -> zonal
			{"Standard_ZRS", "Standard_LRS", true},   // zonal -> non-zonal
			{"Standard_LRS", "Standard_GRS", false},  // both non-zonal
			{"Standard_ZRS", "Standard_GZRS", false}, // both zonal
			{"Premium_LRS", "Premium_ZRS", true},     // tier-agnostic
			{"", "Standard_ZRS", false},              // missing side
			{"Standard_ZRS", "", false},              // missing side
		}
		for _, tc := range cases {
			if got := skuZoneMigration(tc.old, tc.new); got != tc.want {
				t.Errorf("skuZoneMigration(%q, %q) = %v, want %v", tc.old, tc.new, got, tc.want)
			}
		}
	})

	t.Run("accountKindRequiresReplace", func(t *testing.T) {
		cases := []struct {
			old, new string
			want     bool
		}{
			{"StorageV2", "StorageV2", false},   // unchanged
			{"Storage", "StorageV2", false},     // the one allowed in-place migration
			{"StorageV2", "BlobStorage", true},  // any other change
			{"BlobStorage", "StorageV2", false}, // new == StorageV2 allowed
			{"", "StorageV2", false},            // missing old
		}
		for _, tc := range cases {
			if got := accountKindRequiresReplace(tc.old, tc.new); got != tc.want {
				t.Errorf("accountKindRequiresReplace(%q, %q) = %v, want %v", tc.old, tc.new, got, tc.want)
			}
		}
	})

	t.Run("largeFileShareDisabled", func(t *testing.T) {
		cases := []struct {
			old, new string
			want     bool
		}{
			{"Enabled", "Disabled", true},  // turning off
			{"Enabled", "", true},          // clearing
			{"Enabled", "Enabled", false},  // unchanged
			{"Disabled", "Enabled", false}, // turning on
			{"", "Enabled", false},         // never enabled
		}
		for _, tc := range cases {
			if got := largeFileShareDisabled(tc.old, tc.new); got != tc.want {
				t.Errorf("largeFileShareDisabled(%q, %q) = %v, want %v", tc.old, tc.new, got, tc.want)
			}
		}
	})
}
