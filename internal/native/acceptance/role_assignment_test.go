package nativeacc

import "testing"

// TestRoleAssignmentWriteEnabled defends the default-true contract of
// roleAssignmentWriteEnabled: only an explicit false value disables the
// role-assignment-write scenarios; empty, whitespace, and unparseable values
// all keep them enabled.
func TestRoleAssignmentWriteEnabled(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{value: "", want: true},       // empty => default on
		{value: "  ", want: true},     // whitespace trims to empty => default on
		{value: "false", want: false}, // explicit disable
		{value: "0", want: false},     // explicit disable
		{value: "FALSE", want: false}, // explicit disable (case-insensitive ParseBool)
		{value: "true", want: true},   // explicit enable
		{value: "1", want: true},      // explicit enable
		{value: "maybe", want: true},  // unparseable => default on
	}

	for _, tc := range cases {
		t.Run(tc.value, func(t *testing.T) {
			t.Setenv(roleAssignmentWriteEnv, tc.value)
			if got := roleAssignmentWriteEnabled(); got != tc.want {
				t.Fatalf("roleAssignmentWriteEnabled() with %s=%q = %t, want %t",
					roleAssignmentWriteEnv, tc.value, got, tc.want)
			}
		})
	}
}
