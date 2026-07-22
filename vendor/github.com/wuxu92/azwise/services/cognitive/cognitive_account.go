package cognitive

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CognitiveAccount provides resource knowledge for Microsoft.CognitiveServices/accounts.
//
// This ARM type is served by two AzureRM Terraform resources, merged here. Only
// knowledge universally true for every account body is unioned; kind-specific value
// constraints on sub-objects that only one kind sets are safe because they fire only
// when that sub-object is present.
//
// Contributing TF resources:
//   - azurerm_cognitive_account (any kind) — cognitive_account_resource.go
//   - azurerm_ai_services      (kind = AIServices variant) — ai_services_resource.go
//   - azurerm_cognitive_account_customer_managed_key applies the encryption block to
//     this same account body (properties.encryption.*); it has no distinct ARM type —
//     cognitive_account_customer_managed_key_resource.go
//
// Sources:
//   - AzureRM cognitive_account_resource.go schema + CRUD (customer-managed-key,
//     network_acls, network_injection, api_properties expand/flatten)
//   - AzureRM ai_services_resource.go schema (public_network_access, bypass default)
//   - AzureRM cognitive_account_customer_managed_key_resource.go (encryption body)
//   - AzureRM validate/account_name.go (name regex)
//   - Azure SDK cognitive/2026-03-01/cognitiveservicesaccounts models + constants.go
type CognitiveAccount struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CognitiveAccount)(nil)

// NewCognitiveAccount returns a CognitiveAccount knowledge instance.
func NewCognitiveAccount() *CognitiveAccount {
	return &CognitiveAccount{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.CognitiveServices/accounts",
			ApiVersions:  []string{"2026-03-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				// `kind` is only conditionally ForceNew (AzureRM allows switching
				// between OpenAI and AIServices in place, forces new otherwise). Not
				// representable declaratively, so it is documented but not listed here.
				//
				// metrics_advisor_* map to properties.apiProperties.* and are
				// unconditionally ForceNew in cognitive_account. They only exist on
				// MetricsAdvisor-kind accounts, so they never appear on ai_services
				// bodies — safe to union.
				{PropertyPath: "properties.apiProperties.aadClientId"},
				{PropertyPath: "properties.apiProperties.aadTenantId"},
				{PropertyPath: "properties.apiProperties.superUser"},
				{PropertyPath: "properties.apiProperties.websiteName"},
				// custom_subdomain_name is unconditionally ForceNew in ai_services but
				// only conditionally ForceNew (change-from-non-empty) in cognitive_account
				// — left out of the shared declarative ForceNew to avoid over-forcing.
			},
			SoftDelete: true, // delete path purges the soft-deleted account (PurgeSoftDeleteOnDestroy)
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// kind and sku.name are Required in every body (ai_services hardcodes
			// kind = "AIServices"); Sku.Name is a required non-pointer in the SDK.
			RequiredFields: []string{
				"kind",
				"sku.name",
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ──
				// validate.AccountName(): must start alphanumeric, then alphanumeric,
				// period, dash or underscore; effectively >= 2 characters.
				{
					Regex:     `^([a-zA-Z0-9]{1}[a-zA-Z0-9_.-]{1,})$`,
					MinLength: 2,
					Message:   "must start with an alphanumeric character and contain only alphanumeric characters, periods, dashes or underscores",
				},
				// ── kind ──
				// validation.StringInSlice in cognitive_account schema (superset that
				// already includes AIServices used by ai_services).
				{
					PropertyPath: "kind",
					AllowedValues: []string{
						"AIServices", "Academic", "AnomalyDetector",
						"Bing.Autosuggest", "Bing.Autosuggest.v7", "Bing.CustomSearch",
						"Bing.Search", "Bing.Search.v7", "Bing.Speech",
						"Bing.SpellCheck", "Bing.SpellCheck.v7", "CognitiveServices",
						"ComputerVision", "ContentModerator", "ConversationalLanguageUnderstanding",
						"ContentSafety", "CustomSpeech", "CustomVision.Prediction",
						"CustomVision.Training", "Emotion", "Face", "FormRecognizer",
						"ImmersiveReader", "LUIS", "LUIS.Authoring", "MetricsAdvisor",
						"OpenAI", "Personalizer", "QnAMaker", "Recommendations",
						"SpeakerRecognition", "Speech", "SpeechServices", "SpeechTranslation",
						"TextAnalytics", "TextTranslation", "WebLM",
					},
					Message: "must be a valid Cognitive Services account kind",
				},
				// ── sku.name ──
				// validation.StringInSlice; superset of cognitive_account + ai_services.
				// Sku.Name is a free-form string in the SDK (no enum constant).
				{
					PropertyPath: "sku.name",
					AllowedValues: []string{
						"C2", "C3", "C4", "D3", "DC0", "E0", "F0", "F1",
						"P0", "P1", "P2", "S", "S0", "S1", "S2", "S3", "S4", "S5", "S6",
					},
					Message: "must be a valid Cognitive Services SKU name",
				},
				// ── properties.publicNetworkAccess ──
				// ai_services enum; cognitive_account maps its bool to the same ARM enum.
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be Enabled or Disabled",
				},
				// ── properties.networkAcls.defaultAction ──
				// validation.StringInSlice(["Allow","Deny"]).
				{
					PropertyPath:  "properties.networkAcls.defaultAction",
					AllowedValues: []string{"Allow", "Deny"},
					Message:       "must be Allow or Deny",
				},
				// ── properties.networkAcls.bypass ──
				// PossibleValuesForByPassSelection().
				{
					PropertyPath:  "properties.networkAcls.bypass",
					AllowedValues: []string{"AzureServices", "None"},
					Message:       "must be AzureServices or None",
				},
				// ── properties.networkInjections[*].scenario ──
				// cognitive_account restricts to "agent"; full SDK ScenarioType enum.
				{
					PropertyPath:  "properties.networkInjections[*].scenario",
					AllowedValues: []string{"agent", "none"},
					Message:       "must be agent or none",
				},
				// ── properties.networkAcls.ipRules[*].value ──
				// validation.Any(IPv4Address, CIDR).
				{
					PropertyPath: "properties.networkAcls.ipRules[*].value",
					Regex:        `^([0-9]{1,3}\.){3}[0-9]{1,3}(/([0-9]|[1-2][0-9]|3[0-2]))?$`,
					Message:      "must be a valid IPv4 address or CIDR (e.g. 10.0.0.1 or 10.0.0.0/24)",
				},
				// ── UUID fields (validation.IsUUID) ──
				{
					PropertyPath: "properties.apiProperties.aadClientId",
					Regex:        `(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
					Message:      "must be a valid UUID",
				},
				{
					PropertyPath: "properties.apiProperties.aadTenantId",
					Regex:        `(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
					Message:      "must be a valid UUID",
				},
				{
					PropertyPath: "properties.encryption.keyVaultProperties.identityClientId",
					Regex:        `(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
					Message:      "must be a valid UUID",
				},
				{
					PropertyPath: "properties.userOwnedStorage[*].identityClientId",
					Regex:        `(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
					Message:      "must be a valid UUID",
				},
			},
			ArrayRules: []azwise.ArrayRule{
				// network_injection block is MaxItems: 1 → single-element ARM array.
				{
					PropertyPath: "properties.networkInjections",
					MaxItems:     1,
					Message:      "only a single network_injection is supported",
				},
			},
			// custom_question_answering_search_service_key is Sensitive; it maps to
			// properties.apiProperties.qnaAzureSearchEndpointKey. The primary/secondary
			// access keys are returned by ListKeys, not part of the account body.
			SensitiveFields: []string{
				"properties.apiProperties.qnaAzureSearchEndpointKey",
			},
			// properties.endpoint is server-assigned and read-only.
			ComputedFields: []string{
				"properties.endpoint",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},          // both default enabled
				{PropertyPath: "properties.disableLocalAuth", Value: false},                 // local_auth_enabled defaults true
				{PropertyPath: "properties.restrictOutboundNetworkAccess", Value: false},    // outbound_network_access_restricted defaults false
				{PropertyPath: "properties.allowProjectManagement", Value: false},           // project_management_enabled defaults false
			},
			RequiredWith: []azwise.RelationalRule{
				// network_acls requires custom_subdomain_name (both contributors).
				{
					Paths:   []string{"properties.networkAcls", "properties.customSubDomainName"},
					Message: "network_acls requires custom_subdomain_name to be set",
				},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewCognitiveAccount()) }
