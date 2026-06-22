package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Index is the parsed bicep-types manifest (internal/azure/generated/index.json).
// It maps every "<ARMType>@<version>" to the types.json file that defines it and
// is the authoritative source for which API versions exist for a resource type.
//
// Resolving "which API version do we generate" through the manifest (instead of
// scanning directories) has two payoffs: the entry's $ref encodes the exact
// types.json path, so no directory layout is hardcoded; and "latest stable" is
// computed per resource type rather than per namespace. Generation, the
// post-compile validator, and every test that loads a bicep fixture resolve
// through here, so vendoring a newer index.json/types.json rolls the whole
// pipeline forward and no version literal can silently drift.
type Index struct {
	dir       string                // directory holding index.json; $ref paths are relative to it
	resources map[string]indexEntry // "<ARMType>@<version>" -> manifest entry
}

type indexEntry struct {
	Ref string `json:"$ref"` // e.g. "storage/microsoft.storage/2025-06-01/types.json#/29"
}

// LoadIndex reads and parses the bicep-types manifest at path. $ref paths in the
// manifest are interpreted relative to the manifest's directory.
func LoadIndex(path string) (*Index, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw struct {
		Resources map[string]indexEntry `json:"resources"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	if len(raw.Resources) == 0 {
		return nil, fmt.Errorf("%s has no resources", path)
	}
	return &Index{dir: filepath.Dir(path), resources: raw.Resources}, nil
}

// LatestStableVersion returns the newest non-preview API version of armType (a
// bare type such as "Microsoft.Storage/storageAccounts"). Versions are named
// "YYYY-MM-DD" (stable) or "YYYY-MM-DD-preview", so a lexicographic max over the
// non-preview ones yields the latest stable. The "@" anchor keeps child types
// (e.g. storageAccounts/blobServices) from leaking into a parent's resolution.
func (ix *Index) LatestStableVersion(armType string) (string, error) {
	prefix := armType + "@"
	latest := ""
	for key := range ix.resources {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		ver := key[len(prefix):]
		if strings.Contains(ver, "preview") {
			continue
		}
		if ver > latest {
			latest = ver
		}
	}
	if latest == "" {
		return "", fmt.Errorf("no stable API version for %q in index", armType)
	}
	return latest, nil
}

// TypesPath resolves a full "<ARMType>@<version>" tag to the absolute path of the
// types.json that defines it, taken verbatim from the manifest entry's $ref (the
// "#/N" JSON-pointer fragment is stripped). No namespace/directory string is
// reconstructed — the manifest is the source of truth for the path.
func (ix *Index) TypesPath(tag string) (string, error) {
	e, ok := ix.resources[tag]
	if !ok {
		// Descriptor casing should match the manifest, but tolerate drift rather
		// than surface a spurious "not found" for a resource that does exist.
		lower := strings.ToLower(tag)
		for k, v := range ix.resources {
			if strings.ToLower(k) == lower {
				e, ok = v, true
				break
			}
		}
	}
	if !ok {
		return "", fmt.Errorf("resource %q not found in index", tag)
	}
	ref := e.Ref
	if i := strings.IndexByte(ref, '#'); i >= 0 {
		ref = ref[:i]
	}
	return filepath.Join(ix.dir, filepath.FromSlash(ref)), nil
}

// ResolveLatestStable returns the latest-stable tag ("<ARMType>@<version>") and
// the types.json path for armType, both derived from the manifest.
func (ix *Index) ResolveLatestStable(armType string) (tag, typesPath string, err error) {
	ver, err := ix.LatestStableVersion(armType)
	if err != nil {
		return "", "", err
	}
	tag = armType + "@" + ver
	typesPath, err = ix.TypesPath(tag)
	return tag, typesPath, err
}
