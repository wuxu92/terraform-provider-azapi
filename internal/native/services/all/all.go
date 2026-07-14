// Package all blank-imports every generated service package so that importing it
// runs service init() functions: descriptor registration into services.Registry
// and any hand-written runtime hook registration in <resource>_hooks.go. It is a
// separate package (not generated itself) to avoid an import cycle: service
// packages import generated for Register/Descriptor, so generated must not import
// them back.
//
// Add one blank import per service under internal/native/services/<service>.
package all

import (
	_ "github.com/Azure/terraform-provider-azapi/internal/native/services/authorization"
	_ "github.com/Azure/terraform-provider-azapi/internal/native/services/keyvault"
	_ "github.com/Azure/terraform-provider-azapi/internal/native/services/managedidentity"
	_ "github.com/Azure/terraform-provider-azapi/internal/native/services/network"
	_ "github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
	_ "github.com/Azure/terraform-provider-azapi/internal/native/services/storage"
	_ "github.com/Azure/terraform-provider-azapi/internal/native/services/web"
)
