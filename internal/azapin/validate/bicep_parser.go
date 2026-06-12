package validate

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ParseBicepTypes loads a types.json file and returns the BicepType trees
// for all ResourceType entries. This is a minimal parser that produces
// validate.BicepType trees (independent of the generator package).
func ParseBicepTypes(path string) (map[string]*BicepType, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entries []rawEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	w := &bicepWalker{entries: entries, cache: make(map[int]*BicepType)}
	result := make(map[string]*BicepType)

	for _, entry := range entries {
		if entry.Type != "ResourceType" || entry.Body == nil {
			continue
		}
		idx, err := parseRef(entry.Body.Ref)
		if err != nil {
			continue
		}
		body := w.resolve(idx, 0)
		if body != nil {
			result[entry.Name] = body
		}
	}
	return result, nil
}

type rawEntry struct {
	Type       string                    `json:"$type"`
	Name       string                    `json:"name,omitempty"`
	Value      string                    `json:"value,omitempty"`
	Body       *rawRef                   `json:"body,omitempty"`
	Properties map[string]*rawProperty   `json:"properties,omitempty"`
	ItemType   *rawRef                   `json:"itemType,omitempty"`
	Elements   []rawRef                  `json:"elements,omitempty"`
}

type rawRef struct {
	Ref string `json:"$ref"`
}

type rawProperty struct {
	Type        rawRef `json:"type"`
	Flags       int    `json:"flags"`
	Description string `json:"description,omitempty"`
}

type bicepWalker struct {
	entries []rawEntry
	cache   map[int]*BicepType
}

func (w *bicepWalker) resolve(idx int, depth int) *BicepType {
	if depth > 20 || idx < 0 || idx >= len(w.entries) {
		return &BicepType{Kind: BicepKindAny}
	}
	if t, ok := w.cache[idx]; ok {
		return t
	}

	entry := w.entries[idx]
	var t *BicepType

	switch entry.Type {
	case "StringType":
		t = &BicepType{Kind: BicepKindString}
	case "StringLiteralType":
		t = &BicepType{Kind: BicepKindStringLiteral, Name: entry.Value}
	case "IntegerType":
		t = &BicepType{Kind: BicepKindInt}
	case "BooleanType":
		t = &BicepType{Kind: BicepKindBool}
	case "ObjectType":
		t = &BicepType{Kind: BicepKindObject, Name: entry.Name, Properties: make(map[string]*BicepProperty)}
		w.cache[idx] = t
		for name, prop := range entry.Properties {
			propIdx, err := parseRef(prop.Type.Ref)
			if err != nil {
				continue
			}
			resolved := w.resolve(propIdx, depth+1)
			if resolved == nil {
				resolved = &BicepType{Kind: BicepKindAny}
			}
			t.Properties[name] = &BicepProperty{
				Name:  name,
				Type:  resolved,
				Flags: prop.Flags,
			}
		}
		return t
	case "ArrayType":
		t = &BicepType{Kind: BicepKindArray}
		w.cache[idx] = t
		if entry.ItemType != nil {
			itemIdx, err := parseRef(entry.ItemType.Ref)
			if err == nil {
				t.ElementType = w.resolve(itemIdx, depth+1)
			}
		}
		return t
	case "UnionType":
		t = &BicepType{Kind: BicepKindUnion}
		w.cache[idx] = t
		for _, elem := range entry.Elements {
			elemIdx, err := parseRef(elem.Ref)
			if err != nil {
				continue
			}
			resolved := w.resolve(elemIdx, depth+1)
			if resolved != nil {
				t.Elements = append(t.Elements, resolved)
			}
		}
		return t
	default:
		t = &BicepType{Kind: BicepKindAny, Name: entry.Name}
	}

	w.cache[idx] = t
	return t
}

func parseRef(ref string) (int, error) {
	if ref == "" || !strings.HasPrefix(ref, "#/") {
		return -1, fmt.Errorf("invalid ref: %s", ref)
	}
	var idx int
	if _, err := fmt.Sscanf(ref[2:], "%d", &idx); err != nil {
		return -1, err
	}
	return idx, nil
}
