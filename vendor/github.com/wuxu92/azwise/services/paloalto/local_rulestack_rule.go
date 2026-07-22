package paloalto

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LocalRulestackRule provides resource knowledge for
// PaloAltoNetworks.Cloudngfw/localRulestacks/localRules.
//
// Mirrors azurerm_palo_alto_local_rulestack_rule. priority is ForceNew (it is also the
// ID name segment) and validated 1..1000000. AzureRM enforces
// ExactlyOneOf(protocol, protocol_ports) which maps to
// properties.protocol XOR properties.protocolPortList.
//
// The category, source and destination blocks are complex nested structures
// (properties.category / properties.source / properties.destination); their inner
// enums/arrays are not flattened into declarative rules here (documented).
//
// Sources:
//   - terraform-provider-azurerm internal/services/paloalto/palo_alto_local_rulestack_rule_resource.go
//     Arguments() L64-182 (name -> properties.ruleName; rulestack_id ForceNew parent;
//     priority ForceNew Int 1..1000000 -> properties.priority; action enum ->
//     properties.actionType; applications Required -> properties.applications;
//     decryption_rule_type Default None -> properties.decryptionRuleType;
//     enabled Default true -> properties.ruleState; logging_enabled Default false ->
//     properties.enableLogging; negate_destination/negate_source Default false ->
//     properties.negateDestination/negateSource; protocol/protocol_ports ExactlyOneOf
//     -> properties.protocol/protocolPortList; inspection_certificate_id ->
//     properties.inboundInspectionCertificate), Create() L192-303, timeouts 30m/5m/30m/30m
//   - go-azure-sdk resource-manager/paloaltonetworks/2025-10-08/localrulesresources
//     model_ruleentry.go, constants.go (ActionEnum, DecryptionRuleTypeEnum, StateEnum,
//     BooleanEnum)
type LocalRulestackRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LocalRulestackRule)(nil)

// NewLocalRulestackRule returns knowledge for the localRules resource.
func NewLocalRulestackRule() *LocalRulestackRule {
	minPriority := int64(1)
	maxPriority := int64(1000000)

	return &LocalRulestackRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "PaloAltoNetworks.Cloudngfw/localRulestacks/localRules",
			ApiVersions:  []string{"2025-10-08"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.priority"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.ruleName",
				"properties.priority",
				"properties.applications",
				"properties.actionType",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.decryptionRuleType", Value: "None"},
				{PropertyPath: "properties.ruleState", Value: "ENABLED"},
				{PropertyPath: "properties.enableLogging", Value: "DISABLED"},
				{PropertyPath: "properties.negateDestination", Value: "FALSE"},
				{PropertyPath: "properties.negateSource", Value: "FALSE"},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.priority",
					MinValue:     &minPriority,
					MaxValue:     &maxPriority,
					Message:      "priority must be between 1 and 1000000",
				},
			},
			ExactlyOneOf: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.protocol", "properties.protocolPortList"},
					Message: "exactly one of protocol or protocol_ports must be set",
				},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "", // resource name
					Regex:        `^[a-zA-Z0-9]([a-zA-Z0-9-]{0,126}[a-zA-Z0-9])?$`,
					MaxLength:    128,
					Message:      "may only contain alphanumeric characters and dashes, must be 1-128 characters and cannot start or end with a dash",
				},
				{
					PropertyPath:  "properties.actionType",
					AllowedValues: []string{"Allow", "DenyResetBoth", "DenyResetServer", "DenySilent"},
					Message:       "must be one of Allow, DenyResetBoth, DenyResetServer, DenySilent",
				},
				{
					PropertyPath:  "properties.decryptionRuleType",
					AllowedValues: []string{"None", "SSLInboundInspection", "SSLOutboundInspection"},
					Message:       "must be one of None, SSLInboundInspection, SSLOutboundInspection",
				},
				{
					PropertyPath:  "properties.ruleState",
					AllowedValues: []string{"ENABLED", "DISABLED"},
					Message:       "must be one of ENABLED, DISABLED",
				},
				{
					PropertyPath:  "properties.enableLogging",
					AllowedValues: []string{"ENABLED", "DISABLED"},
					Message:       "must be one of ENABLED, DISABLED",
				},
				{
					PropertyPath:  "properties.negateDestination",
					AllowedValues: []string{"TRUE", "FALSE"},
					Message:       "must be one of TRUE, FALSE",
				},
				{
					PropertyPath:  "properties.negateSource",
					AllowedValues: []string{"TRUE", "FALSE"},
					Message:       "must be one of TRUE, FALSE",
				},
			},
		},
	}
}

func init() { azwise.Register(NewLocalRulestackRule()) }
