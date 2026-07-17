package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func runStringValidator(v validator.String, val types.String) *validator.StringResponse {
	req := validator.StringRequest{Path: path.Root("field"), ConfigValue: val}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), req, resp)
	return resp
}

func TestVirtualNetworkBgpCommunity(t *testing.T) {
	cases := []struct {
		name    string
		value   types.String
		wantErr bool
	}{
		{"valid", types.StringValue("12076:20000"), false},
		{"missing colon", types.StringValue("12076"), true},
		{"asn out of range", types.StringValue("70000:20000"), true},
		{"community out of range", types.StringValue("12076:70000"), true},
		{"asn zero", types.StringValue("0:20000"), true},
		{"non-numeric asn", types.StringValue("abc:20000"), true},
		{"null skipped", types.StringNull(), false},
		{"unknown skipped", types.StringUnknown(), false},
	}
	v := VirtualNetworkBgpCommunity()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := runStringValidator(v, tc.value).Diagnostics.HasError(); got != tc.wantErr {
				t.Errorf("HasError() = %v, want %v", got, tc.wantErr)
			}
		})
	}
}
