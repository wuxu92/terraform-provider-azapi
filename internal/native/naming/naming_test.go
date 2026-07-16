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
		// --- AzureRM-authority names (from the generated reference table /
		// overrides): the ARM type maps to a single AzureRM resource, whose noun
		// wins over the mechanical derivation. ---
		{"Microsoft.KeyVault/vaults", "azapi_key_vault"},
		{"Microsoft.DataFactory/factories", "azapi_data_factory"},
		{"Microsoft.Network/virtualNetworks", "azapi_virtual_network"},
		{"Microsoft.Network/virtualNetworks/subnets", "azapi_subnet"},
		{"Microsoft.Network/networkSecurityGroups", "azapi_network_security_group"},
		{"Microsoft.ContainerService/managedClusters", "azapi_kubernetes_cluster"},
		{"Microsoft.Web/serverfarms", "azapi_service_plan"},
		{"Microsoft.Sql/servers", "azapi_mssql_server"},
		{"Microsoft.Sql/servers/databases", "azapi_mssql_database"},
		{"Microsoft.Cache/redis", "azapi_redis_cache"},
		{"Microsoft.EventHub/namespaces", "azapi_eventhub_namespace"},
		{"Microsoft.DBforPostgreSQL/flexibleServers", "azapi_postgresql_flexible_server"},
		{"Microsoft.ManagedIdentity/userAssignedIdentities", "azapi_user_assigned_identity"},
		{"Microsoft.Kusto/clusters", "azapi_kusto_cluster"},
		{"Microsoft.Kusto/clusters/databases", "azapi_kusto_database"},
		{"Microsoft.DocumentDB/databaseAccounts", "azapi_cosmosdb_account"},
		// Overrides pin scope-based ARM types the extractor cannot resolve.
		{"Microsoft.Authorization/roleAssignments", "azapi_role_assignment"},
		{"Microsoft.Authorization/roleDefinitions", "azapi_role_definition"},
		// Third-party providers resolved through the reference table.
		{"Dynatrace.Observability/monitors", "azapi_dynatrace_monitor"},
		{"NewRelic.Observability/monitors", "azapi_new_relic_monitor"},

		// --- Mechanical fallback: ARM types absent from AzureRM or mapping to
		// several AzureRM resources (ambiguous) fall back to service+segment
		// derivation with stutter/dup collapse. ---
		{"Microsoft.Storage/storageAccounts", "azapi_storage_account"},                           // ambiguous in AzureRM
		{"Microsoft.Storage/storageAccounts/blobServices", "azapi_storage_account_blob_service"}, // no standalone AzureRM resource
		{"Microsoft.Resources/resourceGroups", "azapi_resource_group"},                           // provider-less ID, not extracted
		{"Microsoft.KeyVault/vaults/keys", "azapi_key_vault_key"},                                // data-plane ID, not extracted
		{"Microsoft.KeyVault/vaults/secrets", "azapi_key_vault_secret"},                          // data-plane ID, not extracted
		{"Microsoft.Compute/virtualMachines", "azapi_virtual_machine"},                           // ambiguous (linux/windows): pinned to os-agnostic noun
		{"Microsoft.Compute/virtualMachines/extensions", "azapi_compute_virtual_machine_extension"},
		{"Microsoft.Web/sites", "azapi_web_site"}, // ambiguous (web/function/logic apps): pinned to the mechanical web_site noun the native resource carries
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
			name:        "key vault key references its key vault parent",
			armType:     "Microsoft.KeyVault/vaults/keys",
			scopes:      ScopeResourceGroup,
			wantName:    "key_vault_id",
			constrained: true,
			matches:     "/subscriptions/s/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/vault1",
			rejects:     "/subscriptions/s/resourceGroups/rg/providers/Microsoft.KeyVault/keys/key1",
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
			name:     "authorization role assignment uses generic parent for extension multi-scope",
			armType:  "Microsoft.Authorization/roleAssignments",
			scopes:   ScopeTenant | ScopeManagementGroup | ScopeSubscription | ScopeResourceGroup | ScopeExtension,
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

// TestAmbiguousResourceNames verifies the curated pins for ARM types the
// extractor leaves ambiguous (applied via init over the generated table) resolve
// through ResourceName, and that no pin duplicates a name the generated table or
// the overrides already own. The init also rebuilds azurermReferenceLower, so a
// hit here confirms that ordering fix.
func TestAmbiguousResourceNames(t *testing.T) {
	want := map[string]string{
		"Microsoft.Compute/virtualMachines":                         "azapi_virtual_machine",
		"Microsoft.Compute/virtualMachineScaleSets":                 "azapi_virtual_machine_scale_set",
		"Microsoft.Compute/restorePointCollections":                 "azapi_virtual_machine_restore_point_collection",
		"Microsoft.DataFactory/factories/dataflows":                 "azapi_data_factory_data_flow",
		"Microsoft.Insights/webTests":                               "azapi_application_insights_web_test",
		"Microsoft.MachineLearningServices/workspaces":              "azapi_machine_learning_workspace",
		"Microsoft.Network/frontDoorWebApplicationFirewallPolicies": "azapi_cdn_frontdoor_firewall_policy",
		"Microsoft.Network/virtualHubs":                             "azapi_virtual_hub",
		"Microsoft.Network/virtualHubs/bgpConnections":              "azapi_virtual_hub_bgp_connection",
		"Microsoft.PolicyInsights/remediations":                     "azapi_policy_remediation",
		"Microsoft.RecoveryServices/vaults/replicationFabrics":      "azapi_site_recovery_fabric",
		"Microsoft.RecoveryServices/vaults/replicationPolicies":     "azapi_site_recovery_replication_policy",
		"Microsoft.StorageMover/storageMovers/endpoints":            "azapi_storage_mover_endpoint",
		"Microsoft.Web/sites":                                       "azapi_web_site",
		"Microsoft.Web/sites/slots":                                 "azapi_web_site_slot",
		// certificates is corrected via azurermReferenceOverrides from the bare
		// app_service collapse to its primary variant.
		"Microsoft.Web/certificates": "azapi_app_service_certificate",
	}
	for armType, exp := range want {
		if got := ResourceName(armType); got != exp {
			t.Errorf("ResourceName(%q) = %q, want %q", armType, got, exp)
		}
	}

	// No two ARM types in the curated set may resolve to the same azapi noun.
	seen := map[string]string{}
	for armType := range ambiguousResourceNames {
		n := ResourceName(armType)
		if prev, ok := seen[n]; ok {
			t.Errorf("duplicate pinned name %q: %q and %q", n, prev, armType)
		}
		seen[n] = armType
	}
}
