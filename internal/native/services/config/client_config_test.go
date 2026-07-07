package config_test

import (
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/native/services/config"
)

// TestClientConfigSharedInstance pins the rendered contract of the single shared
// config.ClientConfig data-source instance. These are load-bearing HCL fragments that
// Key Vault (and any other resource needing the running identity's tenant/object ids)
// embeds verbatim: the declaration block plus the attribute references. Asserting exact
// string equality means a drift in the shared instance's tfType/label, or in how the
// DataSourceConfigBase renders Config/RefOf/IDRef, reddens the exact row that owns it.
func TestClientConfigSharedInstance(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			name: "Config declares the current client_config data source",
			got:  config.ClientConfig.Config(),
			want: `data "azapi_client_config" "current" {}`,
		},
		{
			name: "RefOf tenant_id",
			got:  config.ClientConfig.RefOf("tenant_id"),
			want: "data.azapi_client_config.current.tenant_id",
		},
		{
			name: "RefOf object_id",
			got:  config.ClientConfig.RefOf("object_id"),
			want: "data.azapi_client_config.current.object_id",
		},
		{
			name: "IDRef",
			got:  config.ClientConfig.IDRef(),
			want: "data.azapi_client_config.current.id",
		},
		{
			name: "DataSourceType",
			got:  config.ClientConfig.DataSourceType(),
			want: "azapi_client_config",
		},
		{
			name: "Label",
			got:  config.ClientConfig.Label(),
			want: "current",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}
