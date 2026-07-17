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

func TestStorageAccountIPRule(t *testing.T) {
	cases := []struct {
		name    string
		value   types.String
		wantErr bool
	}{
		{"public ip", types.StringValue("23.45.1.0"), false},
		{"public cidr", types.StringValue("23.45.1.0/30"), false},
		{"private 10/8", types.StringValue("10.0.0.1"), true},
		{"private 172.16", types.StringValue("172.16.0.1"), true},
		{"private 172.31", types.StringValue("172.31.255.255"), true},
		{"boundary 172.15 public", types.StringValue("172.15.0.1"), false},
		{"boundary 172.32 public", types.StringValue("172.32.0.1"), false},
		{"private 192.168", types.StringValue("192.168.1.1"), true},
		{"boundary 192.167 public", types.StringValue("192.167.1.1"), false},
		{"bad format", types.StringValue("not-an-ip"), true},
		{"prefix above 30", types.StringValue("23.45.1.0/31"), true},
		{"null skipped", types.StringNull(), false},
		{"unknown skipped", types.StringUnknown(), false},
	}
	v := StorageAccountIPRule()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := runStringValidator(v, tc.value).Diagnostics.HasError(); got != tc.wantErr {
				t.Errorf("HasError() = %v, want %v", got, tc.wantErr)
			}
		})
	}
}
