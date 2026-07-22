// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitycenter

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SecurityCenterContact provides resource knowledge for
// Microsoft.Security/securityContacts.
//
// Mirrors azurerm_security_center_contact. Subscription-scoped; name is
// envelope-owned and ForceNew. properties.email is required. The
// alertNotifications / alertsToAdmins shape differs across API versions
// (preview SDK uses string enums On/Off; newer APIs use nested objects), so
// only the universally-present email requirement is encoded.
//
// Sources:
//   - terraform-provider-azurerm internal/services/securitycenter/security_center_contact_resource.go
//     (schema lines 42-71: name ForceNew, email Required; Create/Update/Delete 60m, Read 5m)
//   - terraform-provider-azurerm internal/services/securitycenter/parse/contact.go
//     (ID .../providers/Microsoft.Security/securityContacts/<name>)
//   - azure-sdk-for-go preview/security/mgmt/v3.0/security ContactProperties
//     (json email/phone/alertNotifications/alertsToAdmins). Preview SDK: no ARM API version pinned.
type SecurityCenterContact struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SecurityCenterContact)(nil)

// NewSecurityCenterContact returns knowledge for the securityContacts resource.
func NewSecurityCenterContact() *SecurityCenterContact {
	return &SecurityCenterContact{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Security/securityContacts",
			// Preview SDK (v3.0/security); no ARM API version detected.
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			RequiredFields: []string{
				"properties.email",
			},
		},
	}
}

func init() { azwise.Register(NewSecurityCenterContact()) }
