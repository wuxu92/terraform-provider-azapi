package loadtestservice

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LoadTest provides resource knowledge for Microsoft.LoadTestService/loadTests.
//
// Mirrors azurerm_load_test. The whole encryption block is ForceNew (CMK config
// cannot be changed in place), so a single rule on properties.encryption covers all
// of key_url, identity.type and identity.resourceId.
//
// Sources:
//   - terraform-provider-azurerm internal/services/loadtestservice/load_test_resource.go
//   - Arguments(): name (ForceNew, envelope), location (ForceNew), description
//     (Optional -> properties.description), encryption (ForceNew block ->
//     properties.encryption), encryption.key_url (ForceNew, StringIsNotEmpty ->
//     properties.encryption.keyUrl), encryption.identity.type (ForceNew, enum ->
//     properties.encryption.identity.type), encryption.identity.identity_id
//     (ForceNew, ValidateUserAssignedIdentityID -> properties.encryption.identity.resourceId)
//   - CustomizeDiff()/EnsureEncryptionIdentityIDExistsInIdentity(): a semantic check
//     that the CMK identity_id is also listed under identity.identity_ids — not a
//     single-field constraint, so not expressible as an azwise rule (documented).
//   - Create/Read/Update/Delete timeouts: 30m / 5m / 30m / 30m
//   - go-azure-sdk resource-manager/loadtestservice/2022-12-01/loadtests
//     model_loadtestproperties.go: dataPlaneURI is stripped from the PUT body by
//     LoadTestProperties.MarshalJSON (read-only).
type LoadTest struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LoadTest)(nil)

// NewLoadTest returns knowledge for the loadTests resource.
func NewLoadTest() *LoadTest {
	return &LoadTest{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.LoadTestService/loadTests",
			ApiVersions:  []string{"2022-12-01"},
			ForceNew: []azwise.ForceNewRule{
				// commonschema.Location() is ForceNew; name is envelope.
				{PropertyPath: "location"},
				// The encryption block and every sub-field are ForceNew; comparing the
				// whole object captures any change.
				{PropertyPath: "properties.encryption"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// dataPlaneURI is present in the GET response but stripped from the PUT body
			// by the SDK marshaller (read-only).
			ComputedFields: []string{
				"properties.dataPlaneURI",
			},
			StringRules: []azwise.StringRule{
				// encryption.key_url: validation.StringIsNotEmpty.
				{
					PropertyPath: "properties.encryption.keyUrl",
					MinLength:    1,
					Message:      "encryption keyUrl must not be empty",
				},
				// encryption.identity.type enum (full SDK Type set).
				{
					PropertyPath:  "properties.encryption.identity.type",
					AllowedValues: []string{"SystemAssigned", "UserAssigned"},
					Message:       "encryption identity type must be SystemAssigned or UserAssigned",
				},
			},
		},
	}
}

func init() { azwise.Register(NewLoadTest()) }
