package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ExpressRoutePort provides resource knowledge for
// Microsoft.Network/expressRoutePorts.
//
// Mirrors azurerm_express_route_port. name and resource_group_name live on the
// operational envelope; only body/envelope-path knowledge is encoded here.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/express_route_port_resource.go
//     (schema resourceArmExpressRoutePort + expressRoutePortSchema, Create body lines 215-229,
//     CRUD timeouts)
//   - go-azure-sdk resource-manager/network/2025-01-01/expressrouteports:
//     id_expressrouteport.go (ARM type casing "expressRoutePorts"),
//     model_expressrouteportpropertiesformat.go, constants.go (billingType, encapsulation,
//     macSecCipher enums)
//
// Not encoded (deliberate):
//   - name carries a StringMatch regex validator (semantic) — ported to the customizer.
//   - macsec_cipher (link1/link2) is an enum on the links[*] array element; azwise/azapin
//     cannot lower a rule through an array element, so it is documented, not emitted.
//   - link1/link2 are Optional+Computed; the service always creates the physical link
//     pair, so nothing is required or defaulted for the links themselves.
type ExpressRoutePort struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ExpressRoutePort)(nil)

// NewExpressRoutePort returns knowledge for the expressRoutePorts resource.
func NewExpressRoutePort() *ExpressRoutePort {
	return &ExpressRoutePort{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/expressRoutePorts",
			ApiVersions:  []string{"2025-01-01"},
			// location, peering_location, bandwidth_in_gbps and encapsulation are all
			// ForceNew in AzureRM (schema lines 128-152).
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "properties.peeringLocation"},
				{PropertyPath: "properties.bandwidthInGbps"},
				{PropertyPath: "properties.encapsulation"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 120 * time.Minute,
				Read:   5 * time.Minute,
				Update: 120 * time.Minute,
				Delete: 120 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.encapsulation",
					AllowedValues: []string{"Dot1Q", "QinQ"},
					Message:       "must be Dot1Q or QinQ",
				},
				{
					PropertyPath:  "properties.billingType",
					AllowedValues: []string{"MeteredData", "UnlimitedData"},
					Message:       "must be MeteredData or UnlimitedData",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.bandwidthInGbps",
					MinValue:     azwise.Ptr(int64(1)),
					Message:      "bandwidth_in_gbps must be at least 1",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// billing_type Default:"MeteredData".
				{PropertyPath: "properties.billingType", Value: "MeteredData"},
			},
			// peering_location, bandwidth_in_gbps and encapsulation are Required.
			RequiredFields: []string{
				"properties.peeringLocation",
				"properties.bandwidthInGbps",
				"properties.encapsulation",
			},
		},
	}
}

func init() { azwise.Register(NewExpressRoutePort()) }
