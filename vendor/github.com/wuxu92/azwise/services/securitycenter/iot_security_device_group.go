// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitycenter

import (
	"time"

	"github.com/wuxu92/azwise"
)

// IotSecurityDeviceGroup provides resource knowledge for
// Microsoft.Security/deviceSecurityGroups.
//
// Mirrors azurerm_iot_security_device_group. Scoped to an IoT Hub
// (.../Microsoft.Devices/IotHubs/<hub>/providers/Microsoft.Security/deviceSecurityGroups/<name>).
// name and iothub_id are envelope/scope inputs and ForceNew. The body carries
// properties.allowlistRules[] and properties.timeWindowRules[]; their nested
// enums (range_rule.type) and thresholds are array-element paths, so they are
// left to the schema layer rather than encoded as scalar rules.
//
// Sources:
//   - terraform-provider-azurerm internal/services/securitycenter/iot_security_device_group_resource.go
//     (schema lines 43-153: name + iothub_id Required ForceNew, allow_rule + range_rule blocks;
//     Create/Update/Delete 30m, Read 5m)
//   - terraform-provider-azurerm internal/services/securitycenter/parse/iot_security_device_group.go
//     (ID segment "deviceSecurityGroups")
//   - azure-sdk-for-go preview/security/mgmt/v3.0/security DeviceSecurityGroupProperties
//     (allowlistRules/timeWindowRules). Preview SDK: no ARM API version pinned.
type IotSecurityDeviceGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*IotSecurityDeviceGroup)(nil)

// NewIotSecurityDeviceGroup returns knowledge for the deviceSecurityGroups resource.
func NewIotSecurityDeviceGroup() *IotSecurityDeviceGroup {
	return &IotSecurityDeviceGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Security/deviceSecurityGroups",
			// Preview SDK (v3.0/security); no ARM API version detected.
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewIotSecurityDeviceGroup()) }
