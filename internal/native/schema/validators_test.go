package schema

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

func TestUUID(t *testing.T) {
	cases := []struct {
		name    string
		value   types.String
		wantErr bool
	}{
		{"valid", types.StringValue("12345678-1234-1234-1234-1234567890ab"), false},
		{"uppercase hex", types.StringValue("ABCDEF01-2345-6789-ABCD-EF0123456789"), false},
		{"too short", types.StringValue("1234-1234"), true},
		{"not hex", types.StringValue("zzzzzzzz-1234-1234-1234-1234567890ab"), true},
		{"empty", types.StringValue(""), true},
		{"null skipped", types.StringNull(), false},
		{"unknown skipped", types.StringUnknown(), false},
	}
	v := UUID()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := runStringValidator(v, tc.value).Diagnostics.HasError(); got != tc.wantErr {
				t.Errorf("HasError() = %v, want %v", got, tc.wantErr)
			}
		})
	}
}

func TestAzureResourceID(t *testing.T) {
	cases := []struct {
		name    string
		value   types.String
		wantErr bool
	}{
		{"valid rg-scoped", types.StringValue("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet"), false},
		{"valid subscription", types.StringValue("/subscriptions/00000000-0000-0000-0000-000000000000"), false},
		{"not an id", types.StringValue("not-a-resource-id"), true},
		{"empty", types.StringValue(""), true},
		{"null skipped", types.StringNull(), false},
		{"unknown skipped", types.StringUnknown(), false},
	}
	v := AzureResourceID()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := runStringValidator(v, tc.value).Diagnostics.HasError(); got != tc.wantErr {
				t.Errorf("HasError() = %v, want %v", got, tc.wantErr)
			}
		})
	}
}
