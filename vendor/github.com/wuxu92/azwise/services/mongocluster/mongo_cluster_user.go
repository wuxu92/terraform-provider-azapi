package mongocluster

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MongoClusterUser provides resource knowledge for
// Microsoft.DocumentDB/mongoClusters/users (azurerm_mongo_cluster_user).
//
// The user resource is create/delete only (sdk.Resource, no Update), so every
// body-mappable field is ForceNew.
//
// Sources:
//   - AzureRM internal/services/mongocluster/mongo_cluster_user_resource.go
//     :52-57  (object_id = resource name, IsUUID, ForceNew)
//     :59     (mongo_cluster_id parent reference, ForceNew)
//     :61-69  (identity_provider_type enum, Required, ForceNew)
//     :71-79  (principal_type enum, Required, ForceNew)
//     :81-105 (role block: database + name enum, Required, ForceNew)
//     :143-155 (Create: ARM body mapping identityProvider.type /
//     identityProvider.properties.principalType / roles)
//     :115,170,210 (timeouts: create 30m, read 5m, delete 30m)
//   - go-azure-sdk resource-manager/mongocluster/2025-09-01/users:
//     model_userproperties.go (identityProvider required, roles, provisioningState
//     read-only), model_entraidentityprovider.go (type discriminator),
//     model_entraidentityproviderproperties.go (principalType),
//     model_databaserole.go (db / role array elements),
//     constants.go (IdentityProviderType, EntraPrincipalType, UserRole enums)
//
// Not encoded (deliberate):
//   - object_id (resource name) is validated by validation.IsUUID, a semantic
//     validator that is not a regex/enum/length constraint. It cannot be a
//     declarative StringRule; it belongs in an azapin customizer validator on the
//     name attribute, handled separately from this knowledge file.
//   - role is a Required list whose elements carry database (StringIsNotEmpty) and
//     name (UserRole enum "root"). Both are array-element paths
//     (properties.roles[*].db / properties.roles[*].role) that azwise/azapin cannot
//     lower through an array element, so no per-element rule is emitted. The roles
//     array itself is still ForceNew and Required.
//   - identityProvider is a discriminated union in the SDK; only the
//     MicrosoftEntraID variant is exposed by AzureRM, so the type enum has the
//     single supported value.
type MongoClusterUser struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MongoClusterUser)(nil)

// NewMongoClusterUser returns knowledge for the users resource.
func NewMongoClusterUser() *MongoClusterUser {
	return &MongoClusterUser{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/mongoClusters/users",
			ApiVersions:  []string{"2025-09-01"},
			SoftDelete:   false,
			// Create/delete-only resource: identity provider, principal type and the
			// roles array are all unconditionally ForceNew in AzureRM.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.identityProvider.type"},
				{PropertyPath: "properties.identityProvider.properties.principalType"},
				{PropertyPath: "properties.roles"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.identityProvider.type",
					AllowedValues: []string{"MicrosoftEntraID"},
					Message:       "must be MicrosoftEntraID",
				},
				{
					PropertyPath:  "properties.identityProvider.properties.principalType",
					AllowedValues: []string{"servicePrincipal", "user"},
					Message:       "must be servicePrincipal or user",
				},
			},
			SensitiveFields: []string{},
			// provisioningState is response-only (absent from the create model).
			ComputedFields: []string{
				"properties.provisioningState",
			},
			DefaultValues: []azwise.DefaultValue{},
			RequiredFields: []string{
				"properties.identityProvider",
				"properties.identityProvider.type",
				"properties.identityProvider.properties.principalType",
				"properties.roles",
			},
		},
	}
}

func init() { azwise.Register(NewMongoClusterUser()) }
