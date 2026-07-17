package customizers_test

import (
	"reflect"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/azure"
	"github.com/Azure/terraform-provider-azapi/internal/native/generator/customizers"
	"github.com/Azure/terraform-provider-azapi/internal/native/schema/validators"
	networkvalidators "github.com/Azure/terraform-provider-azapi/internal/native/services/network/validators"
	"github.com/Azure/terraform-provider-azapi/internal/native/typegraph"
)

// virtualNetworkDef loads the real latest-stable virtual network bicep defs from
// the embedded source the provider ships, runs PostProcess + customizers.Apply (so
// the graph reflects exactly what the emitter consumes), and returns the vnet
// definition. Skips when the version/types.json can't be resolved.
func virtualNetworkDef(t *testing.T) *typegraph.ResourceDefinition {
	t.Helper()
	const armType = "Microsoft.Network/virtualNetworks"

	version, err := azure.GetLatestStableApiVersion(armType)
	if err != nil {
		t.Skipf("no stable version for %s: %v", armType, err)
	}
	location, err := azure.GetResourceTypeLocation(armType, version)
	if err != nil {
		t.Skipf("no types.json for %s: %v", armType, err)
	}
	data, err := azure.StaticFiles.ReadFile("generated/" + location)
	if err != nil {
		t.Skipf("types.json not found: %v", err)
	}
	defs, err := typegraph.ParseTypesJSON(data)
	if err != nil {
		t.Fatalf("ParseTypesJSON: %v", err)
	}

	typegraph.PostProcess(defs)
	customizers.Apply(defs)

	for _, def := range defs {
		if typegraph.ARMTypeOf(def) == armType {
			return def
		}
	}
	t.Fatalf("no %s def in parsed types (version %s)", armType, version)
	return nil
}

// TestVirtualNetworkCustomizerRequiresIpamPool guards the customizer contract:
// every ipam_pool_prefix_allocations element must carry a Required pool (and a
// Required pool.id). The ARM IpamPoolPrefixAllocation swagger declares no required
// fields, so the generator lowers pool to Optional+Computed; but ARM unconditionally
// dereferences pool.id at apply ("InvalidResourceId: String  is not a valid resource
// ID" when pool is omitted). The customizer promotes pool to Required on both the
// addressSpace and subnet ipam arrays — which share one bicep element type — so a
// missing pool fails at plan time instead of as a cryptic ARM 400.
func TestVirtualNetworkCustomizerRequiresIpamPool(t *testing.T) {
	def := virtualNetworkDef(t)

	for _, arrayPath := range []string{
		"properties.addressSpace.ipamPoolPrefixAllocations",
		"properties.subnets.properties.ipamPoolPrefixAllocations",
	} {
		pool := typegraph.FindProperty(def, arrayPath+".pool")
		if !pool.Flags.IsRequired() {
			t.Errorf("%s.pool: Flags=%d, want Required (omitting pool must fail at plan time)", arrayPath, pool.Flags)
		}
		id := typegraph.FindProperty(def, arrayPath+".pool.id")
		if !id.Flags.IsRequired() {
			t.Errorf("%s.pool.id: Flags=%d, want Required", arrayPath, id.Flags)
		}
		// number_of_ip_addresses: Required + AzureRM's positive-number regex.
		count := typegraph.FindProperty(def, arrayPath+".numberOfIpAddresses")
		if !count.Flags.IsRequired() {
			t.Errorf("%s.numberOfIpAddresses: Flags=%d, want Required", arrayPath, count.Flags)
		}
		if !hasRegexValidator(count.Validators, `^[1-9]\d*$`) {
			t.Errorf("%s.numberOfIpAddresses missing positive-number regex validator: %#v", arrayPath, count.Validators)
		}
		// The array rejects an empty list (SizeAtLeast(1)) and caps at 2 (SizeAtMost(2),
		// matching AzureRM's MaxItems: 2 on ip_address_pool).
		arr := typegraph.FindProperty(def, arrayPath)
		if !hasListSizeAtLeast(arr.Validators, 1) {
			t.Errorf("%s missing SizeAtLeast(1) validator (empty list must fail at plan time): %#v", arrayPath, arr.Validators)
		}
		if !hasListSizeAtMost(arr.Validators, 2) {
			t.Errorf("%s missing SizeAtMost(2) validator (more than 2 pools must fail at plan time): %#v", arrayPath, arr.Validators)
		}
	}
}

// TestVirtualNetworkCustomizerIpamPoolUnshared proves IsolateArrayElement un-shared
// the two ipam arrays before flagging: both must be Required independently, so the
// mutation on one array's element type did not silently ride a shared pointer (or
// miss the second array). It also confirms the addressSpace and subnet element types
// are distinct pointers after Apply.
func TestVirtualNetworkCustomizerIpamPoolUnshared(t *testing.T) {
	def := virtualNetworkDef(t)

	asElem := typegraph.FindProperty(def, "properties.addressSpace.ipamPoolPrefixAllocations").Type.ElementType
	subnetElem := typegraph.FindProperty(def, "properties.subnets.properties.ipamPoolPrefixAllocations").Type.ElementType
	if asElem == subnetElem {
		t.Fatal("addressSpace and subnet ipam element types are the same pointer; IsolateArrayElement did not un-share them")
	}
	if !asElem.Properties["pool"].Flags.IsRequired() || !subnetElem.Properties["pool"].Flags.IsRequired() {
		t.Fatalf("both ipam pools must be Required: addressSpace=%d subnet=%d",
			asElem.Properties["pool"].Flags, subnetElem.Properties["pool"].Flags)
	}
}

// TestVirtualNetworkCustomizerLeavesBgpAndDdos guards the pre-existing customizer
// rules that must survive alongside the ipam pool change.
func TestVirtualNetworkCustomizerLeavesBgpAndDdos(t *testing.T) {
	def := virtualNetworkDef(t)

	bgp := typegraph.FindProperty(def, "properties.bgpCommunities.virtualNetworkCommunity")
	if !hasValidatorFunc(bgp.Validators, networkvalidators.VirtualNetworkBgpCommunity) {
		t.Errorf("bgpCommunities.virtualNetworkCommunity missing VirtualNetworkBgpCommunity validator: %#v", bgp.Validators)
	}
	ddos := typegraph.FindProperty(def, "properties.ddosProtectionPlan.id")
	if !hasValidatorFunc(ddos.Validators, validators.AzureResourceID) {
		t.Errorf("ddosProtectionPlan.id missing AzureResourceID validator: %#v", ddos.Validators)
	}
}

func hasValidatorFunc(vs []typegraph.DescriptionValidator, fn any) bool {
	want := reflect.ValueOf(fn).Pointer()
	for _, v := range vs {
		if v.Kind == typegraph.ValidatorFunc && v.Func != nil && reflect.ValueOf(v.Func).Pointer() == want {
			return true
		}
	}
	return false
}

func hasRegexValidator(vs []typegraph.DescriptionValidator, pattern string) bool {
	for _, v := range vs {
		if v.Kind == typegraph.ValidatorRegex && v.Pattern == pattern {
			return true
		}
	}
	return false
}

func hasListSizeAtLeast(vs []typegraph.DescriptionValidator, min int64) bool {
	for _, v := range vs {
		if v.Kind == typegraph.ValidatorListSizeAtLeast && v.Min != nil && *v.Min == min {
			return true
		}
	}
	return false
}

func hasListSizeAtMost(vs []typegraph.DescriptionValidator, max int64) bool {
	for _, v := range vs {
		if v.Kind == typegraph.ValidatorListSizeAtMost && v.Max != nil && *v.Max == max {
			return true
		}
	}
	return false
}
