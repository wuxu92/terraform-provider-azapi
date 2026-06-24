package nativeacc

import "fmt"

// Reusable, composable resource config builders. Each returns a complete resource
// block; a resource's dependencies are injected as reference parameters (the
// dependency's Terraform address, e.g. "azapi_resource_group.rg.id"). Inside a
// nested container that address resolves to an ancestor resource already applied to
// the shared workspace, so the dependent reuses it without re-declaring it.

// resourceGroupConfig builds a resource-group block named acctest-rg-<n>.
func resourceGroupConfig(label string) string {
	return resourceGroupConfigNamed(label, "acctest-rg-{{.RandomInteger}}")
}

// resourceGroupConfigNamed builds a resource-group block with an explicit name.
func resourceGroupConfigNamed(label, name string) string {
	return fmt.Sprintf(`
resource "azapi_resource_group" %q {
  name            = %q
  subscription_id = "/subscriptions/{{.SubscriptionID}}"
  location        = "{{.Location}}"
}
`, label, name)
}

// storageAccountConfig builds a StorageV2 account block. rgIDRef is the injected
// resource-group reference; extra is an optional trailing fragment such as a
// properties block (leading newline, 2-space indent).
func storageAccountConfig(label, rgIDRef, sku, extra string) string {
	return fmt.Sprintf(`
resource "azapi_storage_account" %q {
  name              = "acctestsa{{.RandomString}}"
  resource_group_id = %s
  location          = "{{.Location}}"
  kind              = "StorageV2"
  sku = {
    name = %q
  }%s
}
`, label, rgIDRef, sku, extra)
}

// blobServiceConfig builds the blob-service child block. saIDRef is the injected
// storage-account reference; body is an optional trailing fragment.
func blobServiceConfig(label, saIDRef, name, body string) string {
	return fmt.Sprintf(`
resource "azapi_storage_account_blob_service" %q {
  name               = %q
  storage_account_id = %s%s
}
`, label, name, saIDRef, body)
}

// accessTier returns a storage-account properties fragment setting access_tier.
func accessTier(tier string) string {
	return fmt.Sprintf(`
  properties = {
    access_tier = %q
  }`, tier)
}

// changeFeed returns a blob-service properties fragment toggling change feed.
func changeFeed(enabled bool) string {
	return fmt.Sprintf(`
  properties = {
    change_feed = {
      enabled = %t
    }
  }`, enabled)
}

// versioning returns a blob-service properties fragment toggling versioning.
func versioning(enabled bool) string {
	return fmt.Sprintf(`
  properties = {
    is_versioning_enabled = %t
  }`, enabled)
}
