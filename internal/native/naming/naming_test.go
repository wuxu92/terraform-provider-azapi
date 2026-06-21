package naming

import (
	"regexp"
	"testing"
)

func TestCamelToSnake(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"minimumTlsVersion", "minimum_tls_version"},
		{"isHnsEnabled", "is_hns_enabled"},
		{"accessTier", "access_tier"},
		{"storageAccounts", "storage_accounts"},
		{"virtualNetworks", "virtual_networks"},
		{"managedClusters", "managed_clusters"},
		{"networkSecurityGroups", "network_security_groups"},
		{"IPRules", "ip_rules"},
		{"dnsEndpointType", "dns_endpoint_type"},
		{"supportsHttpsTrafficOnly", "supports_https_traffic_only"},
		{"allowBlobPublicAccess", "allow_blob_public_access"},
		{"isNfsV3Enabled", "is_nfs_v3_enabled"},
		{"isSftpEnabled", "is_sftp_enabled"},
		{"keyPolicy", "key_policy"},
		{"", ""},
		{"a", "a"},
		{"ABC", "abc"},
		{"AzureAD", "azure_ad"},
	}

	for _, tt := range tests {
		got := CamelToSnake(tt.input)
		if got != tt.want {
			t.Errorf("CamelToSnake(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSnakeToCamel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"minimum_tls_version", "minimumTlsVersion"},
		{"is_hns_enabled", "isHnsEnabled"},
		{"access_tier", "accessTier"},
		{"storage_accounts", "storageAccounts"},
	}

	for _, tt := range tests {
		got := SnakeToCamel(tt.input)
		if got != tt.want {
			t.Errorf("SnakeToCamel(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestResourceName(t *testing.T) {
	tests := []struct {
		armType string
		want    string
	}{
		{"Microsoft.Storage/storageAccounts", "azapi_storage_account"},
		{"Microsoft.Storage/storageAccounts/blobServices", "azapi_storage_account_blob_service"},
		{"Microsoft.KeyVault/vaults", "azapi_keyvault_vault"},
		{"Microsoft.KeyVault/vaults/keys", "azapi_keyvault_vault_key"},
		{"Microsoft.KeyVault/vaults/secrets", "azapi_keyvault_vault_secret"},
		{"Microsoft.Network/virtualNetworks", "azapi_network_virtual_network"},
		{"Microsoft.Network/virtualNetworks/subnets", "azapi_network_virtual_network_subnet"},
		{"Microsoft.Network/networkSecurityGroups", "azapi_network_security_group"},
		{"Microsoft.Compute/virtualMachines", "azapi_compute_virtual_machine"},
		{"Microsoft.Compute/virtualMachines/extensions", "azapi_compute_virtual_machine_extension"},
		{"Microsoft.ContainerService/managedClusters", "azapi_containerservice_managed_cluster"},
		{"Microsoft.Web/sites", "azapi_web_site"},
		{"Microsoft.Sql/servers", "azapi_sql_server"},
		{"Microsoft.Sql/servers/databases", "azapi_sql_server_database"},
		{"Microsoft.Cache/redis", "azapi_cache_redis"},
		{"Microsoft.EventHub/namespaces", "azapi_eventhub_namespace"},
		{"Microsoft.DBforPostgreSQL/flexibleServers", "azapi_dbforpostgresql_flexible_server"},
		{"Microsoft.ManagedIdentity/userAssignedIdentities", "azapi_managedidentity_user_assigned_identity"},
		// Third-party providers
		{"Dynatrace.Observability/monitors", "azapi_dynatrace_monitor"},
		{"NewRelic.Observability/monitors", "azapi_newrelic_monitor"},
	}

	for _, tt := range tests {
		got := ResourceName(tt.armType)
		if got != tt.want {
			t.Errorf("ResourceName(%q) = %q, want %q", tt.armType, got, tt.want)
		}
	}
}

func TestIsValidTerraformName(t *testing.T) {
	valid := []string{"access_tier", "minimum_tls_version", "is_hns_enabled", "a", "abc_123"}
	for _, name := range valid {
		if !IsValidTerraformName(name) {
			t.Errorf("IsValidTerraformName(%q) = false, want true", name)
		}
	}

	invalid := []string{"minimumTlsVersion", "AccessTier", "123abc", "has-hyphen", "HAS SPACE"}
	for _, name := range invalid {
		if IsValidTerraformName(name) {
			t.Errorf("IsValidTerraformName(%q) = true, want false", name)
		}
	}
}

func TestParentReference(t *testing.T) {
	tests := []struct {
		name        string
		armType     string
		scopes      int
		wantName    string
		constrained bool
		matches     string // an ID the validator must accept (when constrained)
		rejects     string // an ID the validator must reject (when constrained)
	}{
		{
			name:        "resource-group scoped resource",
			armType:     "Microsoft.Storage/storageAccounts",
			scopes:      ScopeResourceGroup,
			wantName:    "resource_group_id",
			constrained: true,
			matches:     "/subscriptions/s/resourceGroups/rg1",
			rejects:     "/subscriptions/s",
		},
		{
			name:        "child resource references its parent type",
			armType:     "Microsoft.Storage/storageAccounts/blobServices",
			scopes:      ScopeResourceGroup,
			wantName:    "storage_account_id",
			constrained: true,
			matches:     "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/acct1",
			rejects:     "/subscriptions/s/resourceGroups/rg",
		},
		{
			name:        "grandchild references its immediate parent type",
			armType:     "Microsoft.Sql/servers/databases",
			scopes:      ScopeResourceGroup,
			wantName:    "server_id",
			constrained: true,
			matches:     "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Sql/servers/srv1",
			rejects:     "/subscriptions/s/resourceGroups/rg",
		},
		{
			// Deeply nested type: the parent ID interleaves type segments with
			// instance names (.../service/<name>/apis/<name>), so the validator
			// must NOT expect the type segments back-to-back.
			name:        "multi-level child interleaves parent type segments and names",
			armType:     "Microsoft.ApiManagement/service/apis/operations",
			scopes:      ScopeResourceGroup,
			wantName:    "api_id",
			constrained: true,
			matches:     "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ApiManagement/service/instance1/apis/api1",
			rejects:     "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ApiManagement/service/instance1",
		},
		{
			name:        "subscription scoped resource",
			armType:     "Microsoft.Authorization/policyDefinitions",
			scopes:      ScopeSubscription,
			wantName:    "subscription_id",
			constrained: true,
			matches:     "/subscriptions/s",
			rejects:     "/subscriptions/s/resourceGroups/rg",
		},
		{
			name:        "management-group scoped resource",
			armType:     "Microsoft.Management/managementGroups/subscriptions",
			scopes:      ScopeManagementGroup,
			wantName:    "management_group_id",
			constrained: true,
			matches:     "/providers/Microsoft.Management/managementGroups/mg1",
			rejects:     "/subscriptions/s",
		},
		{
			name:     "multi-scope falls back to generic parent_id",
			armType:  "Microsoft.Foo/bars",
			scopes:   ScopeResourceGroup | ScopeExtension,
			wantName: "parent_id",
		},
		{
			name:     "unknown scope falls back to generic parent_id",
			armType:  "Microsoft.Foo/bars",
			scopes:   0,
			wantName: "parent_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref := ParentReference(tt.armType, tt.scopes)
			if ref.Name != tt.wantName {
				t.Fatalf("Name = %q, want %q", ref.Name, tt.wantName)
			}
			if tt.constrained {
				if ref.Pattern == "" {
					t.Fatalf("expected a validator pattern for %q", tt.armType)
				}
				re := regexp.MustCompile(ref.Pattern)
				if !re.MatchString(tt.matches) {
					t.Errorf("pattern %q should match %q", ref.Pattern, tt.matches)
				}
				if re.MatchString(tt.rejects) {
					t.Errorf("pattern %q should reject %q", ref.Pattern, tt.rejects)
				}
			} else if ref.Pattern != "" {
				t.Errorf("expected unconstrained parent (empty pattern), got %q", ref.Pattern)
			}
		})
	}
}

func TestTypeReferenceName(t *testing.T) {
	tests := map[string]string{
		"Microsoft.Storage/storageAccounts":          "storage_account",
		"Microsoft.Network/virtualNetworks":          "virtual_network",
		"Microsoft.ContainerService/managedClusters": "managed_cluster",
		"Microsoft.Sql/servers":                      "server",
		"Microsoft.Sql/servers/databases":            "database",
	}
	for in, want := range tests {
		if got := TypeReferenceName(in); got != want {
			t.Errorf("TypeReferenceName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParentARMType(t *testing.T) {
	tests := map[string]string{
		"Microsoft.Storage/storageAccounts/blobServices": "Microsoft.Storage/storageAccounts",
		"Microsoft.Sql/servers/databases":                "Microsoft.Sql/servers",
		"Microsoft.Storage/storageAccounts":              "Microsoft.Storage/storageAccounts",
	}
	for in, want := range tests {
		if got := parentARMType(in); got != want {
			t.Errorf("parentARMType(%q) = %q, want %q", in, got, want)
		}
	}
}
