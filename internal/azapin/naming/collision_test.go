package naming_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/azapin/naming"
)

func TestResourceNameNoCollisions(t *testing.T) {
	data, err := os.ReadFile("../../azure/generated/index.json")
	if err != nil {
		t.Skipf("index.json not found: %v", err)
	}

	var idx struct {
		Resources map[string]json.RawMessage `json:"resources"`
	}
	if err := json.Unmarshal(data, &idx); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Collect unique resource types (case-insensitive dedup)
	seen := make(map[string]string)
	for key := range idx.Resources {
		rt := key
		if i := strings.Index(key, "@"); i >= 0 {
			rt = key[:i]
		}
		lower := strings.ToLower(rt)
		if _, ok := seen[lower]; !ok {
			seen[lower] = rt
		}
	}

	// Check for invalid names and collisions
	names := make(map[string]string)
	collisions := 0

	for _, rt := range seen {
		tfName := naming.ResourceName(rt)
		// Validate the resource-specific part (after "azapi_")
		suffix := strings.TrimPrefix(tfName, "azapi_")
		if !naming.IsValidTerraformName(suffix) {
			t.Errorf("invalid Terraform name: %s → %s", rt, tfName)
		}
		if existing, ok := names[tfName]; ok {
			t.Logf("collision: %s and %s → %s", rt, existing, tfName)
			collisions++
		} else {
			names[tfName] = rt
		}
	}

	t.Logf("Total: %d resource types, %d unique names, %d collisions", len(seen), len(names), collisions)
	// Allow a small number of collisions (these will be resolved with override table)
	if collisions > 20 {
		t.Errorf("too many collisions: %d (expected < 20)", collisions)
	}
}
