package nativeacc

import (
	"os"
	"strconv"
	"strings"

	ginkgo "github.com/onsi/ginkgo/v2"
)

// roleAssignmentWriteEnv toggles the acceptance scenarios that require unrestricted
// Microsoft.Authorization/roleAssignments/write. It defaults to enabled — the canonical
// CI identity carries the role — so set it to a false value only when running under a
// Contributor-class identity that would otherwise fail those scenarios.
const roleAssignmentWriteEnv = "ARM_TEST_ROLE_ASSIGNMENT_WRITE"

// roleAssignmentWriteEnabled reports whether role-assignment-write-backed scenarios
// should run. It defaults to true: an unset or unparseable value enables them, and only
// an explicit false value ("false", "0", …) disables them.
func roleAssignmentWriteEnabled() bool {
	val := strings.TrimSpace(os.Getenv(roleAssignmentWriteEnv))
	if val == "" {
		return true
	}
	enabled, err := strconv.ParseBool(val)
	return err != nil || enabled
}

// SkipIfNoRoleAssignmentWrite skips the current spec when role-assignment-write-backed
// scenarios are disabled. Some Azure operations trigger an implicit role assignment that
// the platform gates behind unrestricted Microsoft.Authorization/roleAssignments/write
// (part of the Owner and User Access Administrator roles) — for example changing a Key
// Vault's permission model between the access-policy and RBAC authorization systems. A
// Contributor-class identity gets a 400 InsufficientPermissions on those calls.
//
// These scenarios run by default (the knob defaults to enabled) because the provider's CI
// identity carries the role. Set ARM_TEST_ROLE_ASSIGNMENT_WRITE=false to skip them when
// running the suite under an identity that lacks it.
func SkipIfNoRoleAssignmentWrite() {
	if !roleAssignmentWriteEnabled() {
		ginkgo.Skip("role-assignment-write scenarios disabled via " + roleAssignmentWriteEnv +
			"=false; they require unrestricted Microsoft.Authorization/roleAssignments/write " +
			"(Owner or User Access Administrator)")
	}
}
