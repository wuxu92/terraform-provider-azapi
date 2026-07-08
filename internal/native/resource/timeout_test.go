package resource

import (
	"testing"
	"time"

	services "github.com/Azure/terraform-provider-azapi/internal/native/services"
)

// TestBaseTimeout defends the runtime contract of Base.timeout: a non-zero
// per-operation field in the baked-in services.Descriptor wins over the
// caller's fallback, while a zero field (or an unknown op) yields the fallback.
func TestBaseTimeout(t *testing.T) {
	// Create/Update are baked in; Read/Delete are left zero to exercise the
	// fallback branch against a partially-populated descriptor.
	baked := &Base{desc: services.Descriptor{Timeouts: services.Timeouts{
		Create: 60 * time.Minute,
		Update: 45 * time.Minute,
	}}}
	// An all-zero Timeouts (empty descriptor) must fall back for every op.
	empty := &Base{desc: services.Descriptor{}}

	cases := []struct {
		name     string
		base     *Base
		op       string
		fallback time.Duration
		want     time.Duration
	}{
		{"create descriptor wins", baked, "create", 30 * time.Minute, 60 * time.Minute},
		{"update descriptor wins", baked, "update", 30 * time.Minute, 45 * time.Minute},
		{"read zero field falls back", baked, "read", 5 * time.Minute, 5 * time.Minute},
		{"delete zero field falls back", baked, "delete", 30 * time.Minute, 30 * time.Minute},
		{"unknown op falls back", baked, "bogus", 7 * time.Minute, 7 * time.Minute},

		{"empty create falls back", empty, "create", 30 * time.Minute, 30 * time.Minute},
		{"empty read falls back", empty, "read", 5 * time.Minute, 5 * time.Minute},
		{"empty update falls back", empty, "update", 45 * time.Minute, 45 * time.Minute},
		{"empty delete falls back", empty, "delete", 10 * time.Minute, 10 * time.Minute},
		{"empty unknown op falls back", empty, "bogus", 7 * time.Minute, 7 * time.Minute},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.base.timeout(tc.op, tc.fallback); got != tc.want {
				t.Fatalf("timeout(%q, %v) = %v, want %v", tc.op, tc.fallback, got, tc.want)
			}
		})
	}
}
