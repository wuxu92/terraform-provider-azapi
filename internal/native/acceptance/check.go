package nativeacc

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	gomega "github.com/onsi/gomega"
)

// Check is a scenario assertion evaluated after an Apply against the resource's
// Terraform state attributes and its existence in Azure.
type Check func(c *checkCtx)

type checkCtx struct {
	attrs  map[string]interface{}
	id     string
	exists func() (bool, error)
}

// Exists asserts the resource exists in Azure (a GET on its ID succeeds).
func Exists() Check {
	return func(c *checkCtx) {
		ok, err := c.exists()
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "checking existence in Azure")
		gomega.Expect(ok).To(gomega.BeTrue(), "resource %s should exist in Azure", c.id)
	}
}

// KeyAssert builds attribute-level assertions for a Terraform state path, e.g.
// Key("properties.access_tier").HasValue("Hot"). Nested objects use dotted
// segments; list/set elements use a numeric segment (e.g. "network_rules.0.action").
type KeyAssert struct{ key string }

// Key starts an attribute assertion for the given dotted Terraform state path.
func Key(key string) KeyAssert { return KeyAssert{key: key} }

// HasValue asserts the attribute equals value (compared as its string form).
func (k KeyAssert) HasValue(value string) Check {
	return func(c *checkCtx) {
		got, ok := lookup(c.attrs, k.key)
		gomega.Expect(ok).To(gomega.BeTrue(), "attribute %q not found in state", k.key)
		gomega.Expect(formatValue(got)).To(gomega.Equal(value), "attribute %q", k.key)
	}
}

// IsSet asserts the attribute is present and non-empty.
func (k KeyAssert) IsSet() Check {
	return func(c *checkCtx) {
		got, ok := lookup(c.attrs, k.key)
		gomega.Expect(ok).To(gomega.BeTrue(), "attribute %q not found in state", k.key)
		gomega.Expect(formatValue(got)).NotTo(gomega.BeEmpty(), "attribute %q should be set", k.key)
	}
}

// Exists asserts the attribute is present in state.
func (k KeyAssert) Exists() Check {
	return func(c *checkCtx) {
		_, ok := lookup(c.attrs, k.key)
		gomega.Expect(ok).To(gomega.BeTrue(), "attribute %q should exist in state", k.key)
	}
}

// lookup navigates a dotted state path through nested maps and lists. A numeric
// segment indexes into a list/set; any other segment keys into a map. Returns the
// value and whether the full path resolved.
func lookup(attrs map[string]interface{}, path string) (interface{}, bool) {
	var cur interface{} = attrs
	for _, seg := range strings.Split(path, ".") {
		switch node := cur.(type) {
		case map[string]interface{}:
			v, ok := node[seg]
			if !ok {
				return nil, false
			}
			cur = v
		case []interface{}:
			idx, err := strconv.Atoi(seg)
			if err != nil || idx < 0 || idx >= len(node) {
				return nil, false
			}
			cur = node[idx]
		default:
			return nil, false
		}
	}
	return cur, true
}

// formatValue renders a tfjson attribute value as a string for comparison.
// tfjson decodes every JSON number as float64, so integral floats are printed
// without a fractional part or scientific notation (e.g. "604800", not "6.048e+05").
func formatValue(v interface{}) string {
	if f, ok := v.(float64); ok {
		if !math.IsInf(f, 0) && !math.IsNaN(f) && f == math.Trunc(f) {
			return strconv.FormatInt(int64(f), 10)
		}
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	return fmt.Sprintf("%v", v)
}
