package nativeacc

import (
	"context"
	"fmt"

	"github.com/Azure/terraform-provider-azapi/internal/clients"
	"github.com/Azure/terraform-provider-azapi/internal/services/parse"
	"github.com/Azure/terraform-provider-azapi/utils"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// existsResource is a generic acceptance.TestResource for any native static
// resource: it reads the ARM resource ID from state and GETs it with the
// resource's API version.
type existsResource struct {
	armType    string
	apiVersion string
}

func newExists(armType, apiVersion string) existsResource {
	return existsResource{armType: armType, apiVersion: apiVersion}
}

func (e existsResource) Exists(ctx context.Context, client *clients.Client, state *terraform.InstanceState) (*bool, error) {
	idStr := state.Attributes["id"]
	if idStr == "" {
		return nil, fmt.Errorf("resource has no id in state")
	}
	id, err := parse.ResourceIDWithResourceType(idStr, e.armType+"@"+e.apiVersion)
	if err != nil {
		return nil, fmt.Errorf("parsing id %q: %w", idStr, err)
	}
	if _, err := client.ResourceClient.Get(ctx, id.AzureResourceId, id.ApiVersion, clients.DefaultRequestOptions()); err != nil {
		if utils.ResponseErrorWasNotFound(err) {
			b := false
			return &b, nil
		}
		return nil, fmt.Errorf("retrieving %s: %w", id.ID(), err)
	}
	b := true
	return &b, nil
}
