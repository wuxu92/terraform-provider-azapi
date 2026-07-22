package digitaltwins

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DigitalTwinsTimeSeriesDatabaseConnection provides resource knowledge for
// Microsoft.DigitalTwins/digitalTwinsInstances/timeSeriesDatabaseConnections
// (azurerm_digital_twins_time_series_database_connection).
//
// The connection body is a discriminated union; AzureRM only implements the
// AzureDataExplorer variant, so all property paths live under
// properties.* of AzureDataExplorerConnectionProperties.
//
// Sources:
//   - AzureRM internal/services/digitaltwins/digital_twins_time_series_database_connection_resource.go
//     :24-115 (schema: every field ForceNew; required vs optional+default)
//     :133-190,192-266 (CRUD timeouts: create 30m, read 5m, delete 30m)
//     :161-176 (expand: TF field -> ARM adx*/eventHub* body paths, defaults)
//   - AzureRM internal/services/digitaltwins/validate/digital_twins_time_series_database_connection_name.go
//     :11-34 (name length 3-50, not all-numeric, regex)
//   - go-azure-sdk resource-manager/digitaltwins/2023-01-31/timeseriesdatabaseconnections:
//     model_azuredataexplorerconnectionproperties.go (ARM body paths + required/optional)
//
// Not encoded (deliberate):
//   - name/digital_twins_id are envelope + parent references, not body properties.
//   - The Update path does not exist (every field ForceNew); replacement is driven
//     by the body-path ForceNew rules below.
type DigitalTwinsTimeSeriesDatabaseConnection struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DigitalTwinsTimeSeriesDatabaseConnection)(nil)

// NewDigitalTwinsTimeSeriesDatabaseConnection returns knowledge for the
// timeSeriesDatabaseConnections resource.
func NewDigitalTwinsTimeSeriesDatabaseConnection() *DigitalTwinsTimeSeriesDatabaseConnection {
	return &DigitalTwinsTimeSeriesDatabaseConnection{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DigitalTwins/digitalTwinsInstances/timeSeriesDatabaseConnections",
			ApiVersions:  []string{"2023-01-31"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// Every body property is ForceNew (AzureRM marks all schema fields ForceNew).
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.adxDatabaseName"},
				{PropertyPath: "properties.adxEndpointUri"},
				{PropertyPath: "properties.adxResourceId"},
				{PropertyPath: "properties.adxTableName"},
				{PropertyPath: "properties.eventHubConsumerGroup"},
				{PropertyPath: "properties.eventHubEndpointUri"},
				{PropertyPath: "properties.eventHubEntityPath"},
				{PropertyPath: "properties.eventHubNamespaceResourceId"},
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath = name attribute).
					Regex:     `^[A-Za-z0-9][A-Za-z0-9-]+[A-Za-z0-9]$`,
					MinLength: 3,
					MaxLength: 50,
					Message:   "must be 3-50 chars, begin and end with a letter or number, contain only letters, numbers, and hyphens, and not be all numeric",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.adxTableName", Value: "AdtPropertyEvents"},
				{PropertyPath: "properties.eventHubConsumerGroup", Value: "$Default"},
			},
			RequiredFields: []string{
				"properties.adxDatabaseName",
				"properties.adxEndpointUri",
				"properties.adxResourceId",
				"properties.eventHubEndpointUri",
				"properties.eventHubEntityPath",
				"properties.eventHubNamespaceResourceId",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDigitalTwinsTimeSeriesDatabaseConnection()) }
