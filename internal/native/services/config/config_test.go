package config_test

import (
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/native/services/config"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/storage"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/web"
)

// Frozen goldens: the exact _Basic Config() output captured from HEAD before the
// EnvelopeRenderer extraction. The refactor is verified byte-identical, so these are
// the load-bearing contract every builder must keep emitting. Any drift in the shared
// renderer or a builder body reddens TestBuilderBasicConfigByteIdentity.
const (
	goldenRG   = "\nresource \"azapi_resource_group\" \"test\" {\n  name            = \"accazapi-rg-{{.RandomInteger}}\"\n  subscription_id = \"/subscriptions/{{.SubscriptionID}}\"\n  location        = \"{{.Location}}\"\n}\n"
	goldenSA   = "\nresource \"azapi_storage_account\" \"test\" {\n  name              = \"acctestsa{{.RandomString}}\"\n  resource_group_id = azapi_resource_group.test.id\n  location          = \"{{.Location}}\"\n  kind              = \"StorageV2\"\n  sku = {\n    name = \"Standard_LRS\"\n  }\n  properties = {}\n}\n"
	goldenBlob = "\nresource \"azapi_storage_account_blob_service\" \"test\" {\n  name               = \"default\"\n  storage_account_id = azapi_storage_account.test.id\n  properties = {\n    change_feed = {\n      enabled = true\n    }\n  }\n}\n"
	goldenSF   = "\nresource \"azapi_service_plan\" \"test\" {\n  name              = \"acctest-asp-{{.RandomString}}\"\n  resource_group_id = azapi_resource_group.test.id\n  location          = \"{{.Location}}\"\n  kind              = \"app\"\n  sku = {\n    name     = \"B1\"\n    capacity = 1\n  }\n  properties = {\n    reserved       = false\n    hyper_v        = false\n    zone_redundant = false\n  }\n}\n"
	goldenWeb  = "\nresource \"azapi_web_site\" \"test\" {\n  name              = \"acctest-web-{{.RandomString}}\"\n  resource_group_id = azapi_resource_group.test.id\n  location          = \"{{.Location}}\"\n  kind              = \"app\"\n  properties = {\n    server_farm_id = azapi_service_plan.test.id\n  }\n}\n"
)

// TestRenderConfigEnvelopeAxes drives ResourceConfigBase.RenderConfig directly, one row
// per envelope axis, asserting the FULL rendered string. Each want is derived by hand
// from the documented rules (2-space indent, key left-padded to len(ParentAttr), %q for
// name/kind/location, raw ParentRef, leading "\nresource", trailing "\n}\n"). A flip in
// any conditional (dropping the location or kind line, quoting ParentRef, mis-padding,
// swapping type/label) reddens the exact case that owns that axis.
func TestRenderConfigEnvelopeAxes(t *testing.T) {
	tests := []struct {
		name string
		base config.ResourceConfigBase
		env  config.ConfigEnvelope
		want string
	}{
		{
			// location ON + kind SET + body appended, resource_group_id parent (width 17).
			name: "location_kind_body_rg_parent",
			base: config.NewResourceConfigBase("azapi_x"),
			env: config.ConfigEnvelope{
				Name:       "n",
				ParentAttr: "resource_group_id",
				ParentRef:  "azapi_resource_group.test.id",
				Location:   true,
				Kind:       "StorageV2",
				Body:       "\n  extra = true",
			},
			want: "\nresource \"azapi_x\" \"test\" {\n  name              = \"n\"\n  resource_group_id = azapi_resource_group.test.id\n  location          = \"{{.Location}}\"\n  kind              = \"StorageV2\"\n  extra = true\n}\n",
		},
		{
			// location ON + kind EMPTY + no body, subscription_id parent with a pre-quoted
			// literal ParentRef emitted raw (width 15). This is the resource-group shape.
			name: "location_nokind_nobody_subscription_prequoted",
			base: config.NewResourceConfigBase("azapi_resource_group"),
			env: config.ConfigEnvelope{
				Name:       "accazapi-rg-{{.RandomInteger}}",
				ParentAttr: "subscription_id",
				ParentRef:  "\"/subscriptions/{{.SubscriptionID}}\"",
				Location:   true,
				Kind:       "",
				Body:       "",
			},
			want: goldenRG,
		},
		{
			// location OFF + kind EMPTY + body appended, storage_account_id parent (width 18).
			// The blob-service singleton shape: proves the location line is omitted entirely.
			name: "nolocation_nokind_body_storage_parent",
			base: config.NewResourceConfigBase("azapi_storage_account_blob_service"),
			env: config.ConfigEnvelope{
				Name:       "default",
				ParentAttr: "storage_account_id",
				ParentRef:  "azapi_storage_account.test.id",
				Location:   false,
				Kind:       "",
				Body:       "\n  properties = {\n    change_feed = {\n      enabled = true\n    }\n  }",
			},
			want: goldenBlob,
		},
		{
			// kind SET but location OFF: guards that the two optional lines are independent —
			// kind still renders while location is suppressed (width 18).
			name: "kind_without_location",
			base: config.NewResourceConfigBase("azapi_x"),
			env: config.ConfigEnvelope{
				Name:       "n",
				ParentAttr: "storage_account_id",
				ParentRef:  "some.ref.id",
				Location:   false,
				Kind:       "app",
				Body:       "",
			},
			want: "\nresource \"azapi_x\" \"test\" {\n  name               = \"n\"\n  storage_account_id = some.ref.id\n  kind               = \"app\"\n}\n",
		},
		{
			// Body axis, empty half: identical to the next row except Body is "".
			name: "empty_body",
			base: config.NewResourceConfigBase("azapi_x"),
			env: config.ConfigEnvelope{
				Name:       "n",
				ParentAttr: "subscription_id",
				ParentRef:  "\"/subscriptions/x\"",
				Location:   false,
				Kind:       "",
				Body:       "",
			},
			want: "\nresource \"azapi_x\" \"test\" {\n  name            = \"n\"\n  subscription_id = \"/subscriptions/x\"\n}\n",
		},
		{
			// Body axis, non-empty half: differs from the row above ONLY in Body, so the want
			// differs by exactly the verbatim-appended fragment.
			name: "nonempty_body",
			base: config.NewResourceConfigBase("azapi_x"),
			env: config.ConfigEnvelope{
				Name:       "n",
				ParentAttr: "subscription_id",
				ParentRef:  "\"/subscriptions/x\"",
				Location:   false,
				Kind:       "",
				Body:       "\n  foo = 1",
			},
			want: "\nresource \"azapi_x\" \"test\" {\n  name            = \"n\"\n  subscription_id = \"/subscriptions/x\"\n  foo = 1\n}\n",
		},
		{
			// Non-default label: proves the `resource %q %q` header uses the builder's label,
			// not the "test" default.
			name: "non_default_label",
			base: config.NewResourceConfigBase("azapi_x", "primary"),
			env: config.ConfigEnvelope{
				Name:       "n",
				ParentAttr: "subscription_id",
				ParentRef:  "\"/subscriptions/x\"",
				Location:   false,
				Kind:       "",
				Body:       "",
			},
			want: "\nresource \"azapi_x\" \"primary\" {\n  name            = \"n\"\n  subscription_id = \"/subscriptions/x\"\n}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.base.RenderConfig(tt.env); got != tt.want {
				t.Errorf("RenderConfig mismatch\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

// TestBuilderBasicConfigByteIdentity is the load-bearing guarantee of the refactor: the
// five static-resource _Basic builders now route through the shared renderer and MUST
// still emit byte-for-byte the frozen goldens captured before extraction. It constructs
// each real builder (with real dependency wiring) and compares its Config() to the golden.
func TestBuilderBasicConfigByteIdentity(t *testing.T) {
	rg := resources.NewResourceGroupCfg()
	sa := storage.NewStorageAccountCfg(rg)
	sf := web.NewWebServerFarmCfg(rg)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"resource_group", resources.ResourceGroupCfg_Basic(rg).Config(), goldenRG},
		{"storage_account", storage.StorageAccountCfg_Basic(sa).Config(), goldenSA},
		{"blob_service", storage.BlobServiceCfg_Basic(storage.NewBlobServiceCfg(sa)).Config(), goldenBlob},
		{"web_server_farm", web.WebServerFarmCfg_Basic(sf).Config(), goldenSF},
		{"web_site", web.WebSiteCfg_Basic(web.NewWebSiteCfg(rg, sf)).Config(), goldenWeb},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s _Basic Config() drifted from frozen golden\n got: %q\nwant: %q", tt.name, tt.got, tt.want)
			}
		})
	}
}
