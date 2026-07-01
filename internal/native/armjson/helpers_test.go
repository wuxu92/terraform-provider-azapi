package armjson

import "testing"

func TestAsMap(t *testing.T) {
	in := map[string]interface{}{"name": "value"}
	if got := AsMap(in); got["name"] != "value" {
		t.Fatalf("AsMap returned %#v", got)
	}
	if got := AsMap(nil); len(got) != 0 {
		t.Fatalf("AsMap(nil) = %#v, want empty map", got)
	}
	if got := AsMap("not a map"); len(got) != 0 {
		t.Fatalf("AsMap(non-map) = %#v, want empty map", got)
	}
}

func TestInt64(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   interface{}
		want int64
	}{
		{name: "int", in: int(42), want: 42},
		{name: "int64", in: int64(42), want: 42},
		{name: "uint", in: uint(42), want: 42},
		{name: "float64", in: float64(42), want: 42},
		{name: "invalid", in: "42", want: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Int64(tc.in); got != tc.want {
				t.Fatalf("Int64(%#v) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}
