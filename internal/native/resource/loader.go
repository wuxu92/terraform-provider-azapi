package resource

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Azure/terraform-provider-azapi/internal/azure"
	"github.com/Azure/terraform-provider-azapi/internal/native/typegraph"
)

// bodyCache memoizes parsed bicep body type graphs per "ARMType@APIVersion".
var (
	bodyCache   = map[string]*typegraph.Type{}
	bodyCacheMu sync.Mutex
)

// loadBody resolves the bicep body type graph for an ARM resource type + API
// version, reading the embedded types.json via the authoritative relative path
// from the schema index (azure.StaticFiles) and applying the same PostProcess
// (cycle-safe, azwise overlay) the generator used. Results are cached.
func loadBody(armType, apiVersion string) (*typegraph.Type, error) {
	key := armType + "@" + apiVersion

	bodyCacheMu.Lock()
	if t, ok := bodyCache[key]; ok {
		bodyCacheMu.Unlock()
		return t, nil
	}
	bodyCacheMu.Unlock()

	location, err := typesJSONLocation(armType, apiVersion)
	if err != nil {
		return nil, err
	}

	data, err := azure.StaticFiles.ReadFile("generated/" + location)
	if err != nil {
		return nil, fmt.Errorf("reading embedded types.json %q: %w", location, err)
	}

	defs, err := typegraph.BuildRuntimeGraph(data)
	if err != nil {
		return nil, fmt.Errorf("building runtime graph for %s: %w", key, err)
	}

	for _, d := range defs {
		if strings.EqualFold(d.Name, key) {
			bodyCacheMu.Lock()
			bodyCache[key] = d.Body
			bodyCacheMu.Unlock()
			return d.Body, nil
		}
	}
	return nil, fmt.Errorf("resource %s not found in %s", key, location)
}

// typesJSONLocation returns the relative path (under generated/) of the types.json
// file containing the given resource type + API version, taken from the schema index.
func typesJSONLocation(armType, apiVersion string) (string, error) {
	s := azure.GetAzureSchema()
	if s == nil {
		return "", fmt.Errorf("azure schema index unavailable")
	}
	for key, res := range s.Resources {
		if !strings.EqualFold(key, armType) {
			continue
		}
		for _, def := range res.Definitions {
			if def.ApiVersion == apiVersion {
				return def.Location.Location, nil
			}
		}
	}
	return "", fmt.Errorf("no types.json location for %s@%s", armType, apiVersion)
}
