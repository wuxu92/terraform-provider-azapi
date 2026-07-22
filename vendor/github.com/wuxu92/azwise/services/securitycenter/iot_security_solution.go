// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitycenter

import (
	"time"

	"github.com/wuxu92/azwise"
)

// IotSecuritySolution provides resource knowledge for
// Microsoft.Security/iotSecuritySolutions.
//
// Mirrors azurerm_iot_security_solution. Resource-group scoped; name and
// location are envelope-owned and ForceNew (location encoded below as a body
// top-level path). Body carries properties.{displayName,iotHubs,status,...}.
//
// Sources:
//   - terraform-provider-azurerm internal/services/securitycenter/iot_security_solution_resource.go
//     (schema lines 55-273: name ForceNew (IotSecuritySolutionName regex), location ForceNew,
//     display_name + iothub_ids Required; create body lines 311-322; Create/Update/Delete 30m, Read 5m)
//   - terraform-provider-azurerm internal/services/securitycenter/validate/iot_security_solution_name.go
//     (name regex ^([-a-zA-Z0-9_.])+$)
//   - azure-sdk-for-go preview/security/mgmt/v3.0/security IoTSecuritySolutionProperties
//     (json displayName/iotHubs/status). Preview SDK: no ARM API version pinned.
type IotSecuritySolution struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*IotSecuritySolution)(nil)

// NewIotSecuritySolution returns knowledge for the iotSecuritySolutions resource.
func NewIotSecuritySolution() *IotSecuritySolution {
	return &IotSecuritySolution{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Security/iotSecuritySolutions",
			// Preview SDK (v3.0/security); no ARM API version detected.
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
			},
			RequiredFields: []string{
				"properties.displayName",
				"properties.iotHubs",
			},
			StringRules: []azwise.StringRule{
				{
					// resource name attribute (PropertyPath == "")
					Regex:   `^([-a-zA-Z0-9_.])+$`,
					Message: "name can only contain letters, digits, '-', '.' or '_'",
				},
			},
		},
	}
}

func init() { azwise.Register(NewIotSecuritySolution()) }
