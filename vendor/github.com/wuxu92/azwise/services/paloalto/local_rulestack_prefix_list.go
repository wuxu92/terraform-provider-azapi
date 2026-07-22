package paloalto

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LocalRulestackPrefixList provides resource knowledge for
// PaloAltoNetworks.Cloudngfw/localRulestacks/prefixlists.
//
// Mirrors azurerm_palo_alto_local_rulestack_prefix_list. properties.prefixList is a
// required, non-empty list of CIDR strings (MinItems 1). The element-level IsCIDR
// check is per-item and is not expressible as a declarative azwise rule (skipped).
//
// Sources:
//   - terraform-provider-azurerm internal/services/paloalto/palo_alto_local_rulestack_prefix_list_resource.go
//     Arguments() L47-82 (name not ForceNew; rulestack_id ForceNew parent;
//     prefix_list Required MinItems 1 (IsCIDR) -> properties.prefixList;
//     audit_comment -> properties.auditComment; description -> properties.description),
//     Create() L88-, timeouts 30m/5m/30m/30m
//   - go-azure-sdk resource-manager/paloaltonetworks/2025-10-08/prefixlistresources
//     model_prefixobject.go (PrefixList []string required)
type LocalRulestackPrefixList struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LocalRulestackPrefixList)(nil)

// NewLocalRulestackPrefixList returns knowledge for the prefixlists resource.
func NewLocalRulestackPrefixList() *LocalRulestackPrefixList {
	return &LocalRulestackPrefixList{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "PaloAltoNetworks.Cloudngfw/localRulestacks/prefixLists",
			ApiVersions:  []string{"2025-10-08"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.prefixList",
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "", // resource name
					Regex:        `^[a-zA-Z0-9]([a-zA-Z0-9-]{0,126}[a-zA-Z0-9])?$`,
					MaxLength:    128,
					Message:      "may only contain alphanumeric characters and dashes, must be 1-128 characters and cannot start or end with a dash",
				},
			},
		},
	}
}

func init() { azwise.Register(NewLocalRulestackPrefixList()) }
