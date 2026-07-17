package typegraph

import "testing"

func TestLengthValidatorBounds(t *testing.T) {
	both := LengthValidator(3, 24)
	if both.Min == nil || *both.Min != 3 || both.Max == nil || *both.Max != 24 {
		t.Errorf("LengthValidator(3,24) = %+v", both)
	}
	atLeast := LengthValidator(3, -1)
	if atLeast.Min == nil || *atLeast.Min != 3 || atLeast.Max != nil {
		t.Errorf("LengthValidator(3,-1) should be min-only, got %+v", atLeast)
	}
	atMost := LengthValidator(-1, 24)
	if atMost.Min != nil || atMost.Max == nil || *atMost.Max != 24 {
		t.Errorf("LengthValidator(-1,24) should be max-only, got %+v", atMost)
	}
}

// TestIsolateArrayElementUnshares proves that isolating one array's element type
// before mutating it does not leak into a sibling array that shares the same
// deduplicated bicep element type (the ipRules / ipv6Rules case).
func TestIsolateArrayElementUnshares(t *testing.T) {
	// shared is one element type referenced by two array properties.
	shared := &Type{Kind: KindObject, Properties: map[string]*Property{
		"value": {Name: "value", Type: &Type{Kind: KindString}},
	}}
	body := &Type{Kind: KindObject, Properties: map[string]*Property{
		"ipRules":   {Name: "ipRules", Type: &Type{Kind: KindArray, ElementType: shared}},
		"ipv6Rules": {Name: "ipv6Rules", Type: &Type{Kind: KindArray, ElementType: shared}},
	}}
	def := &ResourceDefinition{Name: "Microsoft.Fake/widgets@2024-01-01", Body: body}

	elem := IsolateArrayElement(def, "ipRules")
	if elem == nil {
		t.Fatal("IsolateArrayElement returned nil")
	}
	elem.Properties["value"].Validators = append(elem.Properties["value"].Validators, Validator(func() {}))

	// ipRules.value carries the validator; ipv6Rules.value (still the shared type)
	// does not.
	ipv4 := FindProperty(def, "ipRules.value")
	ipv6 := FindProperty(def, "ipv6Rules.value")
	if len(ipv4.Validators) != 1 {
		t.Errorf("ipRules.value validators = %d, want 1", len(ipv4.Validators))
	}
	if len(ipv6.Validators) != 0 {
		t.Errorf("ipv6Rules.value validators = %d, want 0 (sharing leaked)", len(ipv6.Validators))
	}
	if ipv4 == ipv6 {
		t.Error("ipRules.value and ipv6Rules.value are still the same *Property (not isolated)")
	}
}

func TestFindPropertyPanicsOnUnknownPath(t *testing.T) {
	def := &ResourceDefinition{
		Name: "Microsoft.Fake/widgets@2024-01-01",
		Body: &Type{Kind: KindObject, Properties: map[string]*Property{
			"properties": {Name: "properties", Type: &Type{Kind: KindObject}},
		}},
	}
	defer func() {
		if recover() == nil {
			t.Error("FindProperty on an unknown path did not panic")
		}
	}()
	FindProperty(def, "properties.doesNotExist")
}

func TestIsolateArrayElementPanicsOnNonArray(t *testing.T) {
	def := &ResourceDefinition{
		Name: "Microsoft.Fake/widgets@2024-01-01",
		Body: &Type{Kind: KindObject, Properties: map[string]*Property{
			"scalar": {Name: "scalar", Type: &Type{Kind: KindString}},
		}},
	}
	defer func() {
		if recover() == nil {
			t.Error("IsolateArrayElement on a non-array path did not panic")
		}
	}()
	IsolateArrayElement(def, "scalar")
}
