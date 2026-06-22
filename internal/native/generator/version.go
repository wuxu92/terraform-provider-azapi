package generator

import (
	"fmt"
	"os"
	"strings"
)

// LatestStableVersion returns the newest non-preview API version directory under
// dir (a "<service>/<namespace>" types root, e.g.
// internal/azure/generated/storage/microsoft.storage). API version directories
// are named "YYYY-MM-DD" (stable) or "YYYY-MM-DD-preview", so a lexicographic max
// over the non-preview names yields the latest stable version.
//
// This is the single source of truth for "which API version do we generate":
// generation and every test that loads a bicep fixture resolve through it, so
// vendoring a newer types.json rolls the whole pipeline forward and no literal
// version can silently drift.
func LatestStableVersion(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	latest := ""
	for _, e := range entries {
		if !e.IsDir() || strings.Contains(e.Name(), "preview") {
			continue
		}
		if e.Name() > latest {
			latest = e.Name()
		}
	}
	if latest == "" {
		return "", fmt.Errorf("no stable API version directory under %s", dir)
	}
	return latest, nil
}
