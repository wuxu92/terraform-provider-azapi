package generator

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
