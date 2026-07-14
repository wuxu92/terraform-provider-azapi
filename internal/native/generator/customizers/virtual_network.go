package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/typegraph"

// customizeVirtualNetwork applies virtual-network-specific schema rules that the
// bicep type graph and azwise overlay cannot express:
//   - bgp_community (properties.bgpCommunities.virtualNetworkCommunity) must be in
//     AzureRM's "asn:community" notation with each value in (0, 65535). Ported as a
//     resource-specific validator in services/network/validators.
//   - ddos_protection_plan.id (properties.ddosProtectionPlan.id) is an ARM resource
//     ID (AzureRM ddosprotectionplans.ValidateDdosProtectionPlanID); the generic
//     AzureResourceID shared validator covers the ID shape.
//   - An ipam_pool_prefix_allocations element must reference a pool and specify how
//     many addresses to allocate. The ARM IpamPoolPrefixAllocation swagger declares
//     no required fields, so the generator lowers both to Optional+Computed; but ARM
//     unconditionally dereferences pool.id at apply ("InvalidResourceId: String  is
//     not a valid resource ID" when pool is omitted), and AzureRM
//     (virtual_network_resource.go) marks both number_of_ip_addresses and pool.id
//     Required. Mirror that here: promote pool to Required (pool.id is already
//     Required via single-optional promotion) and number_of_ip_addresses to Required
//     with AzureRM's positive-number regex, so a missing pool or a non-numeric count
//     fails at plan time instead of as a cryptic ARM 400. The array itself is
//     Optional+Computed (one side of the address_space/ipam ExactlyOneOf), so an
//     empty list would otherwise satisfy "set"; SizeAtLeast(1) rejects an explicit
//     empty list, and SizeAtMost(2) caps it at a dual-stack IPv4/IPv6 pool pair
//     (matching AzureRM's MaxItems: 2 on ip_address_pool). Both the addressSpace
//     and subnet ipam
//     arrays share one bicep element type; IsolateArrayElement un-shares each before
//     flagging so the rules land independently.
func customizeVirtualNetwork(def *typegraph.ResourceDefinition) {
	bgp := typegraph.FindProperty(def, "properties.bgpCommunities.virtualNetworkCommunity")
	bgp.Validators = append(bgp.Validators, typegraph.CustomValidator("VirtualNetworkBgpCommunity()"))

	ddosID := typegraph.FindProperty(def, "properties.ddosProtectionPlan.id")
	ddosID.Validators = append(ddosID.Validators, typegraph.SharedValidator("AzureResourceID()"))

	for _, path := range []string{
		"properties.addressSpace.ipamPoolPrefixAllocations",
		"properties.subnets.properties.ipamPoolPrefixAllocations",
	} {
		elem := typegraph.IsolateArrayElement(def, path)
		pool := elem.Properties["pool"]
		if pool == nil {
			panic("native: customizeVirtualNetwork: " + path + " element has no pool property")
		}
		pool.Flags |= typegraph.FlagRequired

		count := elem.Properties["numberOfIpAddresses"]
		if count == nil {
			panic("native: customizeVirtualNetwork: " + path + " element has no numberOfIpAddresses property")
		}
		count.Flags |= typegraph.FlagRequired
		count.Validators = append(count.Validators, typegraph.RegexValidator(
			`^[1-9]\d*$`,
			"number_of_ip_addresses must be a string representing a positive number",
		))

		arr := typegraph.FindProperty(def, path)
		arr.Validators = append(arr.Validators,
			typegraph.ListSizeAtLeastValidator(1),
			typegraph.ListSizeAtMostValidator(2),
		)
	}
}
