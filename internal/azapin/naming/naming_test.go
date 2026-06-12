package naming

import "testing"

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
