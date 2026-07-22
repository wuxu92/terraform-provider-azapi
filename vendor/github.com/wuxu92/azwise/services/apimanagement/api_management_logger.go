package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementLogger provides resource knowledge for
// Microsoft.ApiManagement/service/loggers.
//
// Mirrors azurerm_api_management_logger. The AzureRM `eventhub` and
// `application_insights` blocks are provider convenience surfaces that both
// expand into ARM `properties.loggerType` + `properties.credentials`
// (map[string]string), so they are not represented as single ARM paths. Both
// blocks are ForceNew, hence loggerType/credentials/resourceId are ForceNew.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_logger_resource.go
//     schema (lines 44-172): resource_id/eventhub/application_insights ForceNew;
//     buffered default true; Create body (lines 206-224).
//     Timeouts 30m/5m/30m/30m (lines 37-42).
//   - Microsoft.ApiManagement/service/loggers@2022-08-01 logger.LoggerContractProperties:
//     loggerType (enum applicationInsights|azureEventHub|azureMonitor), description,
//     credentials (map), isBuffered, resourceId.
type ApiManagementLogger struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementLogger)(nil)

// NewApiManagementLogger returns knowledge for the API Management logger resource.
func NewApiManagementLogger() *ApiManagementLogger {
	return &ApiManagementLogger{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/loggers",
			ApiVersions:  []string{"2022-08-01"},
			ForceNew: []azwise.ForceNewRule{
				// resource_id (azure.ValidateResourceID), ForceNew.
				{PropertyPath: "properties.resourceId"},
				// eventhub / application_insights blocks are ForceNew; both drive
				// loggerType + credentials.
				{PropertyPath: "properties.loggerType"},
				{PropertyPath: "properties.credentials"},
			},
			DefaultValues: []azwise.DefaultValue{
				// buffered default true (schema line 165).
				{PropertyPath: "properties.isBuffered", Value: true},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.loggerType",
					AllowedValues: []string{"applicationInsights", "azureEventHub", "azureMonitor"},
					Message:       "loggerType must be one of applicationInsights, azureEventHub, azureMonitor",
				},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementLogger()) }
