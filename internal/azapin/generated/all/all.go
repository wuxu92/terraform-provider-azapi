// Package all blank-imports every generated service package so that importing it
// runs their init()-time registrations and populates generated.Registry. It is a
// separate package (not generated itself) to avoid an import cycle: service
// packages import generated for Register/Descriptor, so generated must not import
// them back.
//
// Add one blank import per service under internal/azapin/generated/<service>.
package all

import (
	_ "github.com/Azure/terraform-provider-azapi/internal/azapin/generated/storage"
)
