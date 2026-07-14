package network

import (
	"github.com/Azure/terraform-provider-azapi/internal/native/services/config"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
)

// VirtualNetworkCfg carries the Terraform address metadata and dependencies for
// azapi_virtual_network acceptance-test scenarios. Construct it with
// NewVirtualNetworkCfg, then wrap it in a scenario type when applying. The parent
// ResourceGroupCfg is held so every scenario renders the same resource_group_id
// reference (e.g. "azapi_resource_group.rg.id"). The returned HCL is a template
// rendered by the acceptance framework ({{.RandomString}}, {{.Location}}).
type VirtualNetworkCfg struct {
	config.ResourceConfigBase
	resourceGroup resources.ResourceGroupCfg
}

// NewVirtualNetworkCfg builds a virtual-network config depending on the parent
// resource-group config: the vnet holds it and references its IDRef as
// resource_group_id. The label is optional — omit it for the single-instance default
// ("test"), or pass an explicit label when a scope holds more than one. The resource
// type is read from the VirtualNetwork descriptor.
func NewVirtualNetworkCfg(resourceGroup resources.ResourceGroupCfg, label ...string) VirtualNetworkCfg {
	return VirtualNetworkCfg{
		ResourceConfigBase: config.NewResourceConfigBase(VirtualNetwork.Name, label...),
		resourceGroup:      resourceGroup,
	}
}

// VirtualNetworkCfg_Basic is a minimal virtual network with a single CIDR address
// space. address_space.address_prefixes satisfies the ExactlyOneOf constraint against
// ipam_pool_prefix_allocations; every other body field rides its azwise default
// (including private_endpoint_vnet_policies = "Disabled").
type VirtualNetworkCfg_Basic VirtualNetworkCfg

func (r VirtualNetworkCfg_Basic) Config() string {
	return VirtualNetworkCfg(r).config(`
  properties = {
    address_space = {
      address_prefixes = ["10.0.0.0/16"]
    }
  }`)
}

// VirtualNetworkCfg_Complete exercises a broad, ARM-valid slice of the virtual
// network body so applying Basic then Complete proves the whole surface survives an
// in-place update (Update -> Read -> empty plan): a widened multi-CIDR address space,
// DHCP DNS servers, the flow-timeout int-range field, the encryption block (enum
// enforcement), and two subnets that between them cover subnet address prefixes, the
// default-outbound-access bool, both private-network-policy enums, a service endpoint
// with an explicit location, and a subnet delegation. DDoS-plan, BGP-community, and
// VM-protection fields are deliberately omitted: each needs a paid or ExpressRoute-
// backed dependency that would make the live scenario non-hermetic.
type VirtualNetworkCfg_Complete VirtualNetworkCfg

func (r VirtualNetworkCfg_Complete) Config() string {
	return VirtualNetworkCfg(r).config(`
  tags = {
    environment = "acctest"
  }
  properties = {
    address_space = {
      address_prefixes = ["10.0.0.0/16", "10.1.0.0/16"]
    }
    dhcp_options = {
      dns_servers = ["10.0.0.4", "10.0.0.5"]
    }
    flow_timeout_in_minutes = 10
    encryption = {
      enabled     = true
      enforcement = "AllowUnencrypted"
    }
    subnets = [
      {
        name = "internal"
        properties = {
          address_prefix                        = "10.0.1.0/24"
          default_outbound_access               = false
          private_endpoint_network_policies     = "Disabled"
          private_link_service_network_policies = "Enabled"
          service_endpoints = [
            {
              service   = "Microsoft.Storage"
              locations = ["*"]
            },
          ]
        }
      },
      {
        name = "delegated"
        properties = {
          address_prefix = "10.0.2.0/24"
          delegations = [
            {
              name = "webapp"
              properties = {
                service_name = "Microsoft.Web/serverFarms"
              }
            },
          ]
        }
      },
    ]
  }`)
}

// VirtualNetworkCfg_Complete_update mutates every in-place-updatable axis of
// Complete so applying Complete then Complete_update proves the whole surface
// survives an Update -> Read -> empty plan: retagged, different DNS servers, a bumped
// flow timeout, both subnet private-network policies flipped, a widened service-
// endpoint set, and a third subnet added to the list. The immutable
// default_outbound_access stays false on the "internal" subnet (ARM fixes it at
// subnet create), and encryption stays enabled/AllowUnencrypted (the only enforcement
// value supported at GA), so the delta is purely mutable-in-place.
type VirtualNetworkCfg_Complete_update VirtualNetworkCfg

func (r VirtualNetworkCfg_Complete_update) Config() string {
	return VirtualNetworkCfg(r).config(`
  tags = {
    environment = "acctest-updated"
    owner       = "netops"
  }
  properties = {
    address_space = {
      address_prefixes = ["10.0.0.0/16", "10.10.0.0/16"]
    }
    dhcp_options = {
      dns_servers = ["10.0.0.6", "10.0.0.7"]
    }
    flow_timeout_in_minutes = 20
    encryption = {
      enabled     = true
      enforcement = "AllowUnencrypted"
    }
    subnets = [
      {
        name = "internal"
        properties = {
          address_prefix                        = "10.0.1.0/24"
          default_outbound_access               = false
          private_endpoint_network_policies     = "Enabled"
          private_link_service_network_policies = "Disabled"
          service_endpoints = [
            {
              service   = "Microsoft.Storage"
              locations = ["*"]
            },
            {
              service   = "Microsoft.KeyVault"
              locations = ["*"]
            },
          ]
        }
      },
      {
        name = "delegated"
        properties = {
          address_prefix = "10.0.2.0/24"
          delegations = [
            {
              name = "webapp"
              properties = {
                service_name = "Microsoft.Web/serverFarms"
              }
            },
          ]
        }
      },
      {
        name = "extra"
        properties = {
          address_prefix = "10.0.3.0/24"
        }
      },
    ]
  }`)
}

func (r VirtualNetworkCfg) config(body string) string {
	return r.RenderConfig(config.ConfigEnvelope{
		Name:       "accazapivnet" + r.ResourceLabel() + "{{.RandomString}}",
		ParentAttr: "resource_group_id",
		ParentRef:  r.resourceGroup.IDRef(),
		Location:   true,
		Body:       body,
	})
}
