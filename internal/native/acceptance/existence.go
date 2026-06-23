package nativeacc

import (
	"fmt"

	"github.com/Azure/terraform-provider-azapi/internal/clients"
	"github.com/Azure/terraform-provider-azapi/internal/services/parse"
	"github.com/Azure/terraform-provider-azapi/utils"
)

// azureExists GETs the ARM resource by its Terraform state ID at the resource's
// API version. It returns (false, nil) when the resource is absent (404) and an
// error for any other failure.
func azureExists(client *clients.Client, armType, apiVersion, idStr string) (bool, error) {
	if idStr == "" {
		return false, fmt.Errorf("resource has no id in state")
	}
	id, err := parse.ResourceIDWithResourceType(idStr, armType+"@"+apiVersion)
	if err != nil {
		return false, fmt.Errorf("parsing id %q: %w", idStr, err)
	}
	if _, err := client.ResourceClient.Get(client.StopContext, id.AzureResourceId, id.ApiVersion, clients.DefaultRequestOptions()); err != nil {
		if utils.ResponseErrorWasNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("retrieving %s: %w", id.ID(), err)
	}
	return true, nil
}
