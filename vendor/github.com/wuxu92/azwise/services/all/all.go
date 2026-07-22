// Package all blank-imports every azwise service knowledge package so that
// importing it runs their init() functions, which self-register each resource's
// knowledge into the azwise registry (azwise.Register).
//
// It is a separate package (not the root azwise package) to avoid an import
// cycle: service packages import the root azwise package for its types and
// Register, so the root must not import them back. Consumers that need the full
// knowledge set registered (the azapi native generator overlay and runtime)
// blank-import this package.
//
// Add one blank import per service package under services/<service>.
package all

import (
	_ "github.com/wuxu92/azwise/services/authorization"
	_ "github.com/wuxu92/azwise/services/datafactory"
	_ "github.com/wuxu92/azwise/services/documentdb"
	_ "github.com/wuxu92/azwise/services/keyvault"
	_ "github.com/wuxu92/azwise/services/kusto"
	_ "github.com/wuxu92/azwise/services/managedidentity"
	_ "github.com/wuxu92/azwise/services/network"
	_ "github.com/wuxu92/azwise/services/resources"
	_ "github.com/wuxu92/azwise/services/storage"
	_ "github.com/wuxu92/azwise/services/web"
)
