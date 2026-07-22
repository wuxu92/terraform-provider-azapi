package cognitive

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CognitiveAccountConnection provides resource knowledge for
// Microsoft.CognitiveServices/accounts/connections.
//
// AzureRM splits this single ARM type into five typed Terraform resources, one per
// auth type. They are merged here; each contributor pins properties.authType to a
// specific value and restricts properties.category to a subset. Only universal
// knowledge is unioned — the authType/category enums use the full ARM SDK sets since
// AzAPI sends raw ARM values.
//
// Contributing TF resources (auth type → source):
//   - azurerm_cognitive_account_connection_api_key                   (ApiKey)                 — cognitive_account_connection_api_key_resource.go
//   - azurerm_cognitive_account_connection_account_key               (AccountKey)             — cognitive_account_connection_account_key_resource.go
//   - azurerm_cognitive_account_connection_custom_keys               (CustomKeys)             — cognitive_account_connection_custom_keys_resource.go
//   - azurerm_cognitive_account_connection_entra_id                  (AAD)                    — cognitive_account_connection_entra_id_resource.go
//   - azurerm_cognitive_account_connection_account_managed_identity  (AccountManagedIdentity) — cognitive_account_connection_account_managed_identity_resource.go
//
// Sources:
//   - AzureRM cognitive_account_connection*.go schemas + Create funcs
//   - AzureRM validate/account_connection_name.go (name regex)
//   - Azure SDK cognitive/2026-03-01/accountconnectionresource models + constants.go
type CognitiveAccountConnection struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CognitiveAccountConnection)(nil)

// NewCognitiveAccountConnection returns a CognitiveAccountConnection knowledge instance.
func NewCognitiveAccountConnection() *CognitiveAccountConnection {
	return &CognitiveAccountConnection{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.CognitiveServices/accounts/connections",
			ApiVersions:  []string{"2026-03-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				// category is ForceNew in every contributor.
				{PropertyPath: "properties.category"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// authType is a required non-pointer in the SDK; category is Required in
			// every contributor's schema.
			RequiredFields: []string{
				"properties.authType",
				"properties.category",
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ──
				// validate.AccountConnectionName(): ^[a-zA-Z0-9][a-zA-Z0-9_-]{2,32}$
				{
					Regex:     `^[a-zA-Z0-9][a-zA-Z0-9_-]{2,32}$`,
					MinLength: 3,
					MaxLength: 33,
					Message:   "must be 3-33 characters, start with an alphanumeric character, and contain only alphanumeric characters, dashes or underscores",
				},
				// ── properties.authType ──
				// PossibleValuesForConnectionAuthType(); each TF resource pins one value.
				{
					PropertyPath: "properties.authType",
					AllowedValues: []string{
						"AAD", "AccessKey", "AccountKey", "AccountManagedIdentity",
						"AgentUserImpersonation", "AgenticIdentityToken", "AgenticUser",
						"ApiKey", "CustomKeys", "DelegatedSAS", "ManagedIdentity", "None",
						"OAuth2", "PAT", "ProjectManagedIdentity", "SAS", "ServicePrincipal",
						"UserEntraToken", "UsernamePassword",
					},
					Message: "must be a valid connection auth type",
				},
				// ── properties.category ──
				// PossibleValuesForConnectionCategory(); union across all contributors.
				{
					PropertyPath: "properties.category",
					AllowedValues: []string{
						"ADLSGen2", "AIServices", "AmazonMws", "AmazonRdsForOracle",
						"AmazonRdsForSqlServer", "AmazonRedshift", "AmazonS3Compatible",
						"ApiKey", "ApiManagement", "AppConfig", "AppInsights", "AzureBlob",
						"AzureContainerAppEnvironment", "AzureDataExplorer",
						"AzureDatabricksDeltaLake", "AzureKeyVault", "AzureMariaDb",
						"AzureMySqlDb", "AzureOneLake", "AzureOpenAI", "AzurePostgresDb",
						"AzureSqlDb", "AzureSqlMi", "AzureStorageAccount",
						"AzureSynapseAnalytics", "AzureTableStorage", "BingLLMSearch",
						"Cassandra", "CognitiveSearch", "CognitiveService", "Concur",
						"ContainerRegistry", "CosmosDb", "CosmosDbMongoDbApi", "Couchbase",
						"CustomKeys", "Databricks", "Db2", "Drill", "Dynamics", "DynamicsAx",
						"DynamicsCrm", "Elasticsearch", "Eloqua", "FileServer", "FtpServer",
						"GenericContainerRegistry", "GenericHttp", "GenericRest", "Git",
						"GoogleAdWords", "GoogleBigQuery", "GoogleCloudStorage", "Greenplum",
						"GroundingWithBingSearch", "GroundingWithCustomSearch", "Hbase",
						"Hdfs", "Hive", "Hubspot", "Impala", "Informix", "Jira", "Magento",
						"ManagedOnlineEndpoint", "MariaDb", "Marketo", "MicrosoftAccess",
						"MicrosoftFabric", "ModelGateway", "MongoDbAtlas", "MongoDbV2",
						"MySql", "Netezza", "ODataRest", "Odbc", "Office365", "OpenAI",
						"Oracle", "OracleCloudStorage", "OracleServiceCloud", "PayPal",
						"Phoenix", "Pinecone", "PostgreSql", "PowerPlatformEnvironment",
						"Presto", "PythonFeed", "QuickBooks", "Redis", "RemoteA2A",
						"RemoteTool", "Responsys", "S3", "Salesforce",
						"SalesforceMarketingCloud", "SalesforceServiceCloud", "SapBw",
						"SapCloudForCustomer", "SapEcc", "SapHana", "SapOpenHub", "SapTable",
						"Serp", "Serverless", "ServiceNow", "Sftp", "SharePointOnlineList",
						"Sharepoint", "Shopify", "Snowflake", "Spark", "SqlServer", "Square",
						"Sybase", "Teradata", "Vertica", "WebTable", "Xero", "Zoho",
					},
					Message: "must be a valid connection category",
				},
			},
			// api_key / account_key store the secret at properties.credentials.key;
			// custom_keys stores a secret map at properties.credentials.keys. These are
			// on the polymorphic credentials object set only by their respective kinds.
			SensitiveFields: []string{
				"properties.credentials.key",
				"properties.credentials.keys",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewCognitiveAccountConnection()) }
