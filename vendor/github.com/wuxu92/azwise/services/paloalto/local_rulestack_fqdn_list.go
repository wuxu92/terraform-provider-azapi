package paloalto

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LocalRulestackFqdnList provides resource knowledge for
// PaloAltoNetworks.Cloudngfw/localRulestacks/fqdnlists.
//
// Mirrors azurerm_palo_alto_local_rulestack_fqdn_list. properties.fqdnList is a
// required, non-empty list of FQDN strings (MinItems 1). The element-level
// StringIsNotEmpty check is per-item and is not expressible as a declarative azwise
// rule (skipped).
//
// Sources:
//   - terraform-provider-azurerm internal/services/paloalto/palo_alto_local_rulestack_fqdn_list_resource.go
//     Arguments() L47-82 (name not ForceNew; rulestack_id ForceNew parent;
//     fully_qualified_domain_names Required MinItems 1 -> properties.fqdnList;
//     audit_comment -> properties.auditComment; description -> properties.description),
//     Create() L88-, timeouts 30m/5m/30m/30m
//   - go-azure-sdk resource-manager/paloaltonetworks/2025-10-08/fqdnlistlocalrulestackresources
//     model_fqdnobject.go (FqdnList []string required)
type LocalRulestackFqdnList struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LocalRulestackFqdnList)(nil)

// NewLocalRulestackFqdnList returns knowledge for the fqdnlists resource.
func NewLocalRulestackFqdnList() *LocalRulestackFqdnList {
	return &LocalRulestackFqdnList{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "PaloAltoNetworks.Cloudngfw/localRulestacks/fqdnLists",
			ApiVersions:  []string{"2025-10-08"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.fqdnList",
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

func init() { azwise.Register(NewLocalRulestackFqdnList()) }
