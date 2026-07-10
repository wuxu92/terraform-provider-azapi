package keyvault

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Azure/terraform-provider-azapi/internal/clients"
	nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"
	"github.com/Azure/terraform-provider-azapi/internal/services/parse"
)

// TestKeyVaultKeyDataPlaneID pins the destroy-time ID translation from ARM state
// to the Key Vault data-plane DELETE target. ARM exposes key_uri values both with
// and without a version segment; destroy must call DELETE on the versionless
// /keys/{name} object using the data-plane API version, not the ARM resource ID.
func TestKeyVaultKeyDataPlaneID(t *testing.T) {
	cases := []struct {
		name   string
		keyURI string
	}{
		{
			name:   "versionless key_uri",
			keyURI: "https://vault.vault.azure.net/keys/my-key",
		},
		{
			name:   "versioned key_uri_with_version",
			keyURI: "https://vault.vault.azure.net/keys/my-key/9a8b7c6d5e4f3210",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := keyVaultKeyDataPlaneID(tc.keyURI, "my-key")
			if err != nil {
				t.Fatalf("keyVaultKeyDataPlaneID returned error: %v", err)
			}
			if got.AzureResourceId != "vault.vault.azure.net/keys/my-key" {
				t.Fatalf("AzureResourceId = %q, want %q", got.AzureResourceId, "vault.vault.azure.net/keys/my-key")
			}
			if got.ApiVersion != "2025-07-01" {
				t.Fatalf("ApiVersion = %q, want %q", got.ApiVersion, "2025-07-01")
			}
		})
	}
}

// TestKeyVaultKeyDataPlaneIDRejectsMismatchedName defends the safety check that
// prevents a stale or corrupted key_uri from deleting a different Key Vault key
// than the ARM resource instance Terraform is destroying.
func TestKeyVaultKeyDataPlaneIDRejectsMismatchedName(t *testing.T) {
	_, err := keyVaultKeyDataPlaneID("https://vault.vault.azure.net/keys/other-key", "my-key")
	if err == nil {
		t.Fatal("keyVaultKeyDataPlaneID succeeded for a key_uri whose name does not match the ARM resource name")
	}
	if !strings.Contains(err.Error(), `points to key "other-key"`) || !strings.Contains(err.Error(), `state is for key "my-key"`) {
		t.Fatalf("error = %q, want it to describe the key_uri/ARM-name mismatch", err.Error())
	}
}

type keyDeleteCred struct{}

func (keyDeleteCred) GetToken(context.Context, policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: "token", ExpiresOn: time.Now().Add(time.Hour)}, nil
}

type keyDeleteTransport struct {
	t         *testing.T
	method    string
	url       string
	callCount int
}

func (t *keyDeleteTransport) Do(req *http.Request) (*http.Response, error) {
	t.callCount++
	t.method = req.Method
	t.url = req.URL.String()
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     http.StatusText(http.StatusOK),
		Body:       io.NopCloser(strings.NewReader(`{}`)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Request:    req,
	}, nil
}

func TestDeleteKeyDataPlaneUsesKeyURI(t *testing.T) {
	transport := &keyDeleteTransport{t: t}
	dp, err := clients.NewDataPlaneClient(keyDeleteCred{}, &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{Transport: transport},
	})
	if err != nil {
		t.Fatalf("NewDataPlaneClient returned error: %v", err)
	}

	diags := diag.Diagnostics{}
	deleteKeyDataPlane(&nativeresource.CrudCtx{
		Ctx:    context.Background(),
		Client: &clients.Client{DataPlaneClient: dp},
		ID: parse.ResourceId{
			AzureResourceId:   "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/kv/keys/my-key",
			ApiVersion:        "2025-05-01",
			AzureResourceType: "Microsoft.KeyVault/vaults/keys",
			Name:              "my-key",
		},
		State: keyVaultKeyState("https://vault.vault.azure.net/keys/my-key/9a8b7c6d5e4f3210"),
		Diags: &diags,
	})

	if diags.HasError() {
		t.Fatalf("deleteKeyDataPlane returned diagnostics: %v", diags.Errors())
	}
	if transport.callCount != 1 {
		t.Fatalf("data-plane calls = %d, want 1", transport.callCount)
	}
	if transport.method != http.MethodDelete {
		t.Fatalf("method = %q, want DELETE", transport.method)
	}
	if !strings.HasPrefix(transport.url, "https://vault.vault.azure.net/keys/my-key?") {
		t.Fatalf("url = %q, want versionless key DELETE URL", transport.url)
	}
	if !strings.Contains(transport.url, "api-version=2025-07-01") {
		t.Fatalf("url = %q, want data-plane api-version", transport.url)
	}
}

func keyVaultKeyState(keyURI string) types.Object {
	propTypes := map[string]attr.Type{
		"key_uri": types.StringType,
	}
	props := types.ObjectValueMust(propTypes, map[string]attr.Value{
		"key_uri": types.StringValue(keyURI),
	})
	return types.ObjectValueMust(map[string]attr.Type{
		"properties": props.Type(context.Background()),
	}, map[string]attr.Value{
		"properties": props,
	})
}
