package azure_test

import (
	"strings"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/azure"
)

func Test_GetAzureSchema(t *testing.T) {
	if azure.GetAzureSchema() == nil {
		t.Errorf("failed to load azure schema")
	}
}

func Test_GetApiVersions(t *testing.T) {
	case1 := "Microsoft.MachineLearningServices/workspaces/computes"
	if len(azure.GetApiVersions(case1)) == 0 {
		t.Errorf("expect multiple api-version but got 0 for Microsoft.MachineLearningServices/workspaces/computes")
	}

	case2 := "Microsoft.MachineLearningServices/workspaces/computes0"
	if len(azure.GetApiVersions(case2)) != 0 {
		t.Errorf("expect 0 api-version but got multiple for Microsoft.MachineLearningServices/workspaces/computes0")
	}
}

func Test_GetResourceDefinition(t *testing.T) {
	case1 := "Microsoft.MachineLearningServices/workspaces/computes"
	versions := azure.GetApiVersions(case1)
	for _, v := range versions {
		def, err := azure.GetResourceDefinition(case1, v)
		if err != nil {
			t.Error(err)
		}
		if def == nil {
			t.Errorf("failed to load resource definition for %s api-version %s", case1, v)
		}
	}
}

func Test_AllBicepTypes(t *testing.T) {
	if schema := azure.GetAzureSchema(); schema == nil {
		t.Fatal("failed to load azure schema")
	} else {
		resources := schema.Resources
		if len(resources) == 0 {
			t.Fatal("expect resources are not empty")
		}
		for resourceName, res := range resources {
			if len(resourceName) == 0 {
				t.Fatal("expect resource name is not empty")
			}
			if res == nil {
				t.Fatalf("expect resource definition is not nil, resource name: %s", resourceName)
			} else {
				definitions := res.Definitions
				if len(definitions) == 0 {
					t.Fatalf("expect resource definitions are not empty, resource name: %s", resourceName)
				}
				for _, definition := range definitions {
					_, err := definition.GetDefinition()
					if err != nil {
						t.Fatal(err)
					}
				}
			}
		}
	}
}

func Test_GetLatestStableApiVersion(t *testing.T) {
	const resourceType = "Microsoft.Storage/storageAccounts"
	got, err := azure.GetLatestStableApiVersion(resourceType)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "preview") {
		t.Errorf("GetLatestStableApiVersion = %q, want a non-preview version", got)
	}
	// Must equal the newest non-preview version GetApiVersions reports.
	want := ""
	for _, v := range azure.GetApiVersions(resourceType) {
		if !strings.Contains(v, "preview") && v > want {
			want = v
		}
	}
	if got != want {
		t.Errorf("GetLatestStableApiVersion = %q, want %q (newest non-preview)", got, want)
	}

	// A resource type absent from the index has no versions, so no stable one.
	if _, err := azure.GetLatestStableApiVersion("Microsoft.Storage/storageAccounts0"); err == nil {
		t.Error("expected an error for a resource type absent from the index")
	}
}

func Test_GetResourceTypeLocation(t *testing.T) {
	const resourceType = "Microsoft.Storage/storageAccounts"
	version, err := azure.GetLatestStableApiVersion(resourceType)
	if err != nil {
		t.Fatal(err)
	}
	location, err := azure.GetResourceTypeLocation(resourceType, version)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(location, "types.json") {
		t.Errorf("GetResourceTypeLocation = %q, want a path ending in types.json", location)
	}
	// The location must resolve to a readable file in the embedded schema.
	if _, err := azure.StaticFiles.ReadFile("generated/" + location); err != nil {
		t.Errorf("location %q not readable from embedded schema: %v", location, err)
	}
	// An api-version absent from the index resolves to no location.
	if _, err := azure.GetResourceTypeLocation(resourceType, "1999-01-01"); err == nil {
		t.Error("expected an error for an api-version absent from the index")
	}
}
