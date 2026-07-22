package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementBackend provides resource knowledge for
// Microsoft.ApiManagement/service/backends.
//
// Mirrors azurerm_api_management_backend. `name` (ForceNew), `api_management_name`
// and `resource_group_name` are envelope references. Top-level ARM body fields:
//   - protocol    -> properties.protocol (required, enum http|soap)
//   - url         -> properties.url (required)
//   - description -> properties.description (1..2000 chars)
//   - resource_id -> properties.resourceId (1..2000 chars)
//   - title       -> properties.title (1..300 chars)
//
// Nested blocks:
//   - credentials            -> properties.credentials {authorization{parameter,scheme},
//                               certificate[], header{}, query{}}
//   - proxy                  -> properties.proxy {url (required), username, password (sensitive)}
//   - tls                    -> properties.tls {validateCertificateChain, validateCertificateName}
//   - service_fabric_cluster -> properties.properties.serviceFabricCluster
//   - circuit_breaker_rule   -> properties.circuitBreaker.rules[*]
//
// Uses the 2024-05-01 SDK (backend_resource.go imports .../2024-05-01/backend).
//
// TODO: the following AzureRM validators target array/nested-block elements that
// the declarative rule types cannot address and are intentionally not emitted:
//   - circuit_breaker_rule.failure_condition.count (IntBetween 1..10000),
//     .percentage (IntBetween 1..100), .status_code_range.min/max (IntBetween
//     200..599) live under properties.circuitBreaker.rules[*] (array elements).
//   - circuit_breaker_rule.name (StringMatch alnum/hyphen) and trip_duration /
//     failure_condition.interval_duration (ISO8601Duration) are also array-element.
//   - credentials.* and tls.* AtLeastOneOf constraints operate inside the single
//     block instances and are not top-level cross-property relations.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_backend_resource.go
//     schema (lines 47-343); Create body (lines 379-409); expand funcs (lines 492-651);
//     Timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/apimanagement/2024-05-01/backend
//     BackendContractProperties: protocol (enum http|soap), url, description,
//     resourceId, title, credentials, proxy, tls, circuitBreaker,
//     properties.serviceFabricCluster.
type ApiManagementBackend struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementBackend)(nil)

// NewApiManagementBackend returns knowledge for the backends resource.
func NewApiManagementBackend() *ApiManagementBackend {
	return &ApiManagementBackend{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/backends",
			ApiVersions:  []string{"2024-05-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.protocol",
				"properties.url",
			},
			SensitiveFields: []string{
				"properties.proxy.password",
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.protocol",
					AllowedValues: []string{"http", "soap"},
					Message:       "protocol must be one of http or soap",
				},
				{PropertyPath: "properties.description", MinLength: 1, MaxLength: 2000, Message: "description must be 1..2000 characters"},
				{PropertyPath: "properties.resourceId", MinLength: 1, MaxLength: 2000, Message: "resource_id must be 1..2000 characters"},
				{PropertyPath: "properties.title", MinLength: 1, MaxLength: 300, Message: "title must be 1..300 characters"},
			},
			// AzureRM: service_fabric_cluster.server_certificate_thumbprints
			// ConflictsWith server_x509_name.
			ConflictsWith: []azwise.RelationalRule{
				{
					Paths: []string{
						"properties.properties.serviceFabricCluster.serverCertificateThumbprints",
						"properties.properties.serviceFabricCluster.serverX509Names",
					},
					Message: "`server_certificate_thumbprints` conflicts with `server_x509_name`",
				},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementBackend()) }
