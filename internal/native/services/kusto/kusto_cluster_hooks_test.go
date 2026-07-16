package kusto

import "testing"

// TestKustoSkuConstraints locks the SKU-name -> (tier, capacity bounds) mapping the
// ValidateConfig hook enforces: the two Dev(No SLA) names are single-instance Basic
// clusters; every Standard-prefixed name is a Standard cluster scaling from 2 to 1000.
// The cases cover both dev SKUs and a representative spread of Standard families
// (including the Standard_L8s_v3 that produced the reported ARM capacity error).
func TestKustoSkuConstraints(t *testing.T) {
	cases := []struct {
		name     string
		wantTier string
		wantMin  int64
		wantMax  int64
	}{
		{"Dev(No SLA)_Standard_D11_v2", "Basic", 1, 1},
		{"Dev(No SLA)_Standard_E2a_v4", "Basic", 1, 1},
		{"Standard_D11_v2", "Standard", 2, 1000},
		{"Standard_L8s_v3", "Standard", 2, 1000},
		{"Standard_E80ids_v4", "Standard", 2, 1000},
		{"Standard_DS14_v2+4TB_PS", "Standard", 2, 1000},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tier, min, max := kustoSkuConstraints(tc.name)
			if tier != tc.wantTier {
				t.Errorf("tier: got %q, want %q", tier, tc.wantTier)
			}
			if min != tc.wantMin || max != tc.wantMax {
				t.Errorf("capacity bounds: got [%d, %d], want [%d, %d]", min, max, tc.wantMin, tc.wantMax)
			}
		})
	}
}
