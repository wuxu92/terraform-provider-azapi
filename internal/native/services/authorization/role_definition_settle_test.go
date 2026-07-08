package authorization

import (
	"testing"
	"time"
)

// rd builds an ARM GET response of the shape roleDefinitionSettled reads:
// {"properties": {"createdOn": <string>, "updatedOn": <string>}}. Both keys are
// always present; absence cases construct the response map directly in the table.
func rd(createdOn, updatedOn string) map[string]interface{} {
	return map[string]interface{}{
		"properties": map[string]interface{}{
			"createdOn": createdOn,
			"updatedOn": updatedOn,
		},
	}
}

// TestRoleDefinitionSettled pins the post-update eventual-consistency boundary of
// roleDefinitionSettled: a response is settled ONLY when the swap has completed
// (createdOn no longer equals the update time) AND the read reflects our write
// (updatedOn is not before the update time). Every branch and both time boundaries
// around updateRequest are covered, plus the absent/empty/unparseable => false paths.
func TestRoleDefinitionSettled(t *testing.T) {
	const reqStr = "2024-01-02T15:04:05Z"
	req, err := time.Parse(time.RFC3339, reqStr)
	if err != nil {
		t.Fatalf("failed to parse fixed updateRequest %q: %v", reqStr, err)
	}

	const (
		earlierCreate = "2023-06-01T00:00:00Z" // original create date, well before req
		staleUpdated  = "2023-12-31T00:00:00Z" // updatedOn of a stale pre-update record
		afterReq      = "2024-01-02T15:04:06Z" // one second after req
		oneSecBefore  = "2024-01-02T15:04:04Z" // one second before req (boundary)
	)

	cases := []struct {
		name string
		resp map[string]interface{}
		want bool
	}{
		{
			// createdOn still equals updateRequest: the freshly written shadow record
			// has not yet been swapped back to the original create date.
			name: "unconsolidated shadow (createdOn == updatedOn == req)",
			resp: rd(reqStr, reqStr),
			want: false,
		},
		{
			// updatedOn predates updateRequest: the GET returned the pre-update record,
			// so the read does not yet reflect our write.
			name: "stale pre-update record (updatedOn before req)",
			resp: rd(earlierCreate, staleUpdated),
			want: false,
		},
		{
			// Swap completed (createdOn reverted) AND updatedOn carries the update time.
			name: "settled (createdOn reverted, updatedOn == req)",
			resp: rd(earlierCreate, reqStr),
			want: true,
		},
		{
			// Same as settled, but updatedOn is strictly after req: still not before,
			// so still settled.
			name: "settled (updatedOn strictly after req)",
			resp: rd(earlierCreate, afterReq),
			want: true,
		},
		{
			// createdOn reverted, but updatedOn is one second before req: the >= boundary
			// is exclusive below, so before(req) is NOT settled.
			name: "boundary (updatedOn one second before req)",
			resp: rd(earlierCreate, oneSecBefore),
			want: false,
		},
		{
			// createdOn absent: roleDefinitionTime returns !ok => false.
			name: "createdOn absent",
			resp: map[string]interface{}{
				"properties": map[string]interface{}{
					"updatedOn": reqStr,
				},
			},
			want: false,
		},
		{
			// updatedOn absent: roleDefinitionUpdatedOn returns !ok => false.
			name: "updatedOn absent",
			resp: map[string]interface{}{
				"properties": map[string]interface{}{
					"createdOn": earlierCreate,
				},
			},
			want: false,
		},
		{
			// properties absent entirely: neither timestamp can be read => false.
			name: "properties absent",
			resp: map[string]interface{}{},
			want: false,
		},
		{
			// empty-string timestamps parse as absent (raw == "") => false.
			name: "empty-string timestamps",
			resp: rd("", ""),
			want: false,
		},
		{
			// unparseable createdOn: RFC3339 parse fails => false, even though updatedOn
			// is valid and would otherwise satisfy the updatedOn branch.
			name: "unparseable createdOn",
			resp: rd("not-a-time", reqStr),
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := roleDefinitionSettled(tc.resp, req); got != tc.want {
				t.Fatalf("roleDefinitionSettled(%#v, %q) = %t, want %t", tc.resp, reqStr, got, tc.want)
			}
		})
	}
}
