package cdn

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CdnFrontDoorOrigin provides resource knowledge for
// Microsoft.Cdn/profiles/originGroups/origins (Azure Front Door origin).
//
// Contributing Terraform resource: azurerm_cdn_frontdoor_origin.
//
// Sources:
//   - terraform-provider-azurerm internal/services/cdn/cdn_frontdoor_origin_resource.go
//     (schema L48-153, Create body L220-238, expandPrivateLinkSettings L434-474)
//   - internal/services/cdn/validate/front_door_origin_name.go (name regex)
//   - go-azure-sdk resource-manager/cdn/2025-12-01/afdorigins:
//     model_afdoriginproperties.go, model_sharedprivatelinkresourceproperties.go.
//
// Notes:
//   - name & cdn_frontdoor_origin_group_id are envelope/parent-owned (ForceNew).
//   - enabled (bool, default true) maps to properties.enabledState.
//   - private_link.target_type maps to properties.sharedPrivateLinkResource.groupId
//     (free-form string in the SDK; AzureRM restricts to a curated set).
//   - Semantic validators not expressible declaratively (ported at the azapin layer,
//     not here): private_link_target_id -> azure.ValidateResourceID (generic
//     AzureResourceID); origin_host_header -> validation.Any(IsIPv6, IsIPv4,
//     StringIsNotEmpty) composite. private_link is only valid on Premium profiles
//     with certificate_name_check_enabled=true (a cross-resource CustomizeDiff guard).
type CdnFrontDoorOrigin struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CdnFrontDoorOrigin)(nil)

func NewCdnFrontDoorOrigin() *CdnFrontDoorOrigin {
	return &CdnFrontDoorOrigin{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Cdn/profiles/originGroups/origins",
			ApiVersions:  []string{"2025-12-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 4 * time.Hour,
				Read:   5 * time.Minute,
				Update: 4 * time.Hour,
				Delete: 6 * time.Hour,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.FrontDoorOriginName
				{
					Regex:     `^[\da-zA-Z][-\da-zA-Z]{0,88}[\da-zA-Z]$`,
					MinLength: 2,
					MaxLength: 90,
					Message:   "must be 2-90 characters, begin and end with a letter or number, and contain only letters, numbers and hyphens",
				},
				// private_link.request_message → ...requestMessage (StringLenBetween 1-140)
				{
					PropertyPath: "properties.sharedPrivateLinkResource.requestMessage",
					MinLength:    1,
					MaxLength:    140,
				},
				// private_link.target_type → ...groupId (curated set)
				{
					PropertyPath:  "properties.sharedPrivateLinkResource.groupId",
					AllowedValues: []string{"blob", "blob_secondary", "Gateway", "managedEnvironments", "sites", "web", "web_secondary"},
				},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.httpPort", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(65535))},
				{PropertyPath: "properties.httpsPort", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(65535))},
				{PropertyPath: "properties.priority", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(5))},
				{PropertyPath: "properties.weight", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(1000))},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.enabledState", Value: "Enabled"},
				{PropertyPath: "properties.httpPort", Value: float64(80)},
				{PropertyPath: "properties.httpsPort", Value: float64(443)},
				{PropertyPath: "properties.priority", Value: float64(1)},
				{PropertyPath: "properties.weight", Value: float64(500)},
				{PropertyPath: "properties.sharedPrivateLinkResource.requestMessage", Value: "Access request for CDN FrontDoor Private Link Origin"},
			},
			ComputedFields: []string{
				"properties.deploymentStatus",
				"properties.provisioningState",
				"properties.originGroupName",
			},
			RequiredFields: []string{
				"properties.hostName",
				"properties.enforceCertificateNameCheck",
			},
		},
	}
}

func init() { azwise.Register(NewCdnFrontDoorOrigin()) }
