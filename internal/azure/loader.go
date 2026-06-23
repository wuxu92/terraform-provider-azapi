package azure

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/Azure/terraform-provider-azapi/internal/azure/types"
)

var schema *Schema

//go:embed generated
var StaticFiles embed.FS

var mutex = &sync.Mutex{}

func GetAzureSchema() *Schema {
	mutex.Lock()
	defer mutex.Unlock()
	if schema == nil {
		data, err := StaticFiles.ReadFile("generated/index.json")
		if err != nil {
			log.Printf("[ERROR] failed to load schema index: %+v", err)
			return nil
		}
		err = json.Unmarshal(data, &schema)
		if err != nil {
			log.Printf("[ERROR] failed to unmarshal schema index: %+v", err)
			return nil
		}
	}
	return schema
}

// skipApiVersions contains resource type + API version combinations that are present
// in the schema but not yet supported by Azure, causing "NoRegisteredProviderFound" errors at runtime.
// Key: resource type (case-insensitive, stored lowercase), Value: set of API versions to skip.
var skipApiVersions = map[string]map[string]bool{
	"microsoft.keyvault/vaults/keys":    {"2026-02-01": true, "2026-03-01-preview": true},
	"microsoft.keyvault/vaults/secrets": {"2026-02-01": true, "2026-03-01-preview": true},
}

func GetApiVersions(resourceType string) []string {
	azureSchema := GetAzureSchema()
	if azureSchema == nil {
		return []string{}
	}
	skipped := skipApiVersions[strings.ToLower(resourceType)]
	res := make([]string, 0)
	for key, value := range azureSchema.Resources {
		if strings.EqualFold(key, resourceType) {
			for _, v := range value.Definitions {
				if skipped[v.ApiVersion] {
					continue
				}
				res = append(res, v.ApiVersion)
			}
		}
	}
	sort.Strings(res)
	return res
}

// GetLatestStableApiVersion returns the newest non-preview API version available
// for resourceType. API versions are named "YYYY-MM-DD" (stable) or
// "YYYY-MM-DD-preview", so the lexicographic max over the non-preview versions is
// the latest stable one. Versions in skipApiVersions are excluded via GetApiVersions.
func GetLatestStableApiVersion(resourceType string) (string, error) {
	latest := ""
	for _, v := range GetApiVersions(resourceType) {
		if strings.Contains(v, "preview") {
			continue
		}
		if v > latest {
			latest = v
		}
	}
	if latest == "" {
		return "", fmt.Errorf("no stable api-version for resource type %s in azure schema index", resourceType)
	}
	return latest, nil
}

func GetResourceDefinition(resourceType, apiVersion string) (*types.ResourceType, error) {
	azureSchema := GetAzureSchema()
	if azureSchema == nil {
		return nil, fmt.Errorf("failed to load azure schema index")
	}
	for key, value := range azureSchema.Resources {
		if strings.EqualFold(key, resourceType) {
			for _, v := range value.Definitions {
				if v.ApiVersion == apiVersion {
					return v.GetDefinition()
				}
			}
		}
	}
	return nil, fmt.Errorf("failed to find resource type %s api-version %s in azure schema index", resourceType, apiVersion)
}

// GetResourceTypeLocation returns the embedded types.json path (relative to the
// generated/ directory, e.g. "storage/microsoft.storage/2025-06-01/types.json")
// that defines resourceType@apiVersion. It is the bicep-types source the native
// generator parses; read it with StaticFiles.ReadFile("generated/" + location).
func GetResourceTypeLocation(resourceType, apiVersion string) (string, error) {
	azureSchema := GetAzureSchema()
	if azureSchema == nil {
		return "", fmt.Errorf("failed to load azure schema index")
	}
	for key, value := range azureSchema.Resources {
		if strings.EqualFold(key, resourceType) {
			for _, v := range value.Definitions {
				if v.ApiVersion == apiVersion {
					return v.Location.Location, nil
				}
			}
		}
	}
	return "", fmt.Errorf("failed to find resource type %s api-version %s in azure schema index", resourceType, apiVersion)
}
