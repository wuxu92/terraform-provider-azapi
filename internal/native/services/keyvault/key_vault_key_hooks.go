package keyvault

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Azure/terraform-provider-azapi/internal/clients"
	nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"
	"github.com/Azure/terraform-provider-azapi/internal/services/parse"
	"github.com/Azure/terraform-provider-azapi/utils"
)

const keyVaultKeyDataPlaneResourceType = "Microsoft.KeyVault/vaults/keys@2025-07-01"

func init() {
	nativeresource.RegisterHooks(KeyVaultKey.Name, &nativeresource.Hooks{
		Delete: deleteKeyDataPlane,
	})
}

// deleteKeyDataPlane deletes a management-plane-created key through Key Vault's
// data-plane endpoint. Microsoft.KeyVault/vaults/keys supports ARM PUT/GET but the
// ARM swagger exposes no DELETE operation, and Azure returns DeleteNotSupported for
// a direct ARM DELETE. Terraform destroy still needs to remove the key object, so
// this hook uses the computed key_uri from the ARM read and calls DELETE
// {vaultBaseUrl}/keys/{key-name}. The result is a normal Key Vault soft-delete; purge
// remains a separate privileged operation.
func deleteKeyDataPlane(c *nativeresource.CrudCtx) {
	if c.Client.DataPlaneClient == nil {
		c.Diags.AddError("Failed to delete Key Vault key", "provider data-plane client is not configured")
		return
	}

	props := nativeresource.AttrObject(c.State, "properties")
	keyURI := nativeresource.AttrString(props, "key_uri")
	if keyURI == "" {
		keyURI = nativeresource.AttrString(props, "key_uri_with_version")
	}

	id, err := keyVaultKeyDataPlaneID(keyURI, c.ID.Name)
	if err != nil {
		c.Diags.AddError("Failed to delete Key Vault key", err.Error())
		return
	}

	if _, err := c.Client.DataPlaneClient.DeleteThenPoll(c.Ctx, id, clients.DefaultRequestOptions()); err != nil && !utils.ResponseErrorWasNotFound(err) {
		c.Diags.AddError("Failed to delete Key Vault key", fmt.Errorf("deleting %s through Key Vault data plane: %w", c.ID.ID(), err).Error())
	}
}

func keyVaultKeyDataPlaneID(keyURI string, armName string) (parse.DataPlaneResourceId, error) {
	if keyURI == "" {
		return parse.DataPlaneResourceId{}, fmt.Errorf("state is missing properties.key_uri; cannot locate Key Vault data-plane endpoint for %q", armName)
	}

	parsed, err := url.Parse(keyURI)
	if err != nil {
		return parse.DataPlaneResourceId{}, fmt.Errorf("parsing properties.key_uri %q: %w", keyURI, err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return parse.DataPlaneResourceId{}, fmt.Errorf("properties.key_uri %q must be an absolute Key Vault URI", keyURI)
	}

	parts := strings.Split(strings.Trim(parsed.EscapedPath(), "/"), "/")
	keySegment := -1
	for i, part := range parts {
		if strings.EqualFold(part, "keys") {
			keySegment = i
			break
		}
	}
	if keySegment == -1 || keySegment+1 >= len(parts) {
		return parse.DataPlaneResourceId{}, fmt.Errorf("properties.key_uri %q must contain /keys/{key-name}", keyURI)
	}

	uriName, err := url.PathUnescape(parts[keySegment+1])
	if err != nil {
		return parse.DataPlaneResourceId{}, fmt.Errorf("parsing key name from properties.key_uri %q: %w", keyURI, err)
	}
	if uriName != armName {
		return parse.DataPlaneResourceId{}, fmt.Errorf("properties.key_uri %q points to key %q, but Terraform state is for key %q", keyURI, uriName, armName)
	}

	return parse.NewDataPlaneResourceId(armName, parsed.Host, keyVaultKeyDataPlaneResourceType)
}
