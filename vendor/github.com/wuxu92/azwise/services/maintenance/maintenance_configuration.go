package maintenance

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MaintenanceConfiguration provides resource knowledge for
// Microsoft.Maintenance/maintenanceConfigurations.
//
// Mirrors azurerm_maintenance_configuration.
//
// Sources:
//   - terraform-provider-azurerm internal/services/maintenance/maintenance_configuration_resource.go
//     schema (lines 55-237: name/location/resource_group_name ForceNew, scope Required enum,
//     visibility enum default Custom, window {duration regex, time_zone enum}, install_patches
//     {reboot enum, linux/windows classifications enums}, in_guest_user_patch_mode enum,
//     properties map), Create (241-303: body Location/Tags + properties.maintenanceScope/
//     visibility/namespace/maintenanceWindow/extensionProperties/installPatches),
//     Read (363-414); timeouts 30m/5m/30m/30m.
//   - internal/services/maintenance/validate/maintenance.go (MaintenanceTimeZone enum).
//   - go-azure-sdk resource-manager/maintenance/2023-04-01/maintenanceconfigurations
//     MaintenanceConfigurationProperties (maintenanceScope/visibility/namespace/
//     maintenanceWindow/installPatches/extensionProperties), MaintenanceWindow
//     (duration/timeZone/...), InputPatchConfiguration (rebootSetting), constants.go
//     (PossibleValuesForMaintenanceScope/RebootOptions/Visibility).
type MaintenanceConfiguration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MaintenanceConfiguration)(nil)

// NewMaintenanceConfiguration returns knowledge for the maintenanceConfigurations resource.
func NewMaintenanceConfiguration() *MaintenanceConfiguration {
	return &MaintenanceConfiguration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Maintenance/maintenanceConfigurations",
			ApiVersions:  []string{"2023-04-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				// name/location/resource_group_name are ForceNew (envelope).
				{PropertyPath: "name"},
				{PropertyPath: "location"},
			},
			// scope (Required:true) -> properties.maintenanceScope.
			RequiredFields: []string{
				"properties.maintenanceScope",
			},
			// visibility is Optional+Default "Custom".
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.visibility", Value: "Custom"},
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). AzureRM StringIsNotEmpty.
					MinLength: 1,
					Message:   "maintenance configuration name must not be empty",
				},
				{
					// scope enum. Full ARM SDK set (PossibleValuesForMaintenanceScope);
					// AzureRM restricts to a subset but AzAPI sends raw ARM values.
					PropertyPath: "properties.maintenanceScope",
					AllowedValues: []string{
						"Extension",
						"Host",
						"InGuestPatch",
						"OSImage",
						"Resource",
						"SQLDB",
						"SQLManagedInstance",
					},
					Message: "scope must be one of Extension, Host, InGuestPatch, OSImage, Resource, SQLDB, SQLManagedInstance",
				},
				{
					// visibility enum (PossibleValuesForVisibility). AzureRM only allows
					// Custom (Public rejected by API) but the ARM SDK set includes Public.
					PropertyPath: "properties.visibility",
					AllowedValues: []string{
						"Custom",
						"Public",
					},
					Message: "visibility must be Custom or Public",
				},
				{
					// window.duration: HH:mm.
					PropertyPath: "properties.maintenanceWindow.duration",
					Regex:        `^(0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]$`,
					Message:      "duration must match the format HH:mm",
				},
				{
					// window.time_zone: TimeZoneInfo system time zones (validate.MaintenanceTimeZone).
					PropertyPath: "properties.maintenanceWindow.timeZone",
					AllowedValues: []string{
						"Afghanistan Standard Time", "Alaskan Standard Time", "Aleutian Standard Time",
						"Altai Standard Time", "Arab Standard Time", "Arabian Standard Time",
						"Arabic Standard Time", "Argentina Standard Time", "Astrakhan Standard Time",
						"Atlantic Standard Time", "AUS Central Standard Time", "Aus Central W. Standard Time",
						"AUS Eastern Standard Time", "Azerbaijan Standard Time", "Azores Standard Time",
						"Bahia Standard Time", "Bangladesh Standard Time", "Belarus Standard Time",
						"Bougainville Standard Time", "Canada Central Standard Time", "Cape Verde Standard Time",
						"Caucasus Standard Time", "Cen. Australia Standard Time", "Central America Standard Time",
						"Central Asia Standard Time", "Central Brazilian Standard Time", "Central Europe Standard Time",
						"Central European Standard Time", "Central Pacific Standard Time", "Central Standard Time",
						"Central Standard Time (Mexico)", "Chatham Islands Standard Time", "China Standard Time",
						"Cuba Standard Time", "Dateline Standard Time", "E. Africa Standard Time",
						"E. Australia Standard Time", "E. Europe Standard Time", "E. South America Standard Time",
						"Easter Island Standard Time", "Eastern Standard Time", "Eastern Standard Time (Mexico)",
						"Egypt Standard Time", "Ekaterinburg Standard Time", "Fiji Standard Time",
						"FLE Standard Time", "Georgian Standard Time", "GMT Standard Time",
						"Greenland Standard Time", "Greenwich Standard Time", "GTB Standard Time",
						"Haiti Standard Time", "Hawaiian Standard Time", "India Standard Time",
						"Iran Standard Time", "Israel Standard Time", "Jordan Standard Time",
						"Kaliningrad Standard Time", "Kamchatka Standard Time", "Korea Standard Time",
						"Libya Standard Time", "Line Islands Standard Time", "Lord Howe Standard Time",
						"Magadan Standard Time", "Magallanes Standard Time", "Marquesas Standard Time",
						"Mauritius Standard Time", "Mid-Atlantic Standard Time", "Middle East Standard Time",
						"Montevideo Standard Time", "Morocco Standard Time", "Mountain Standard Time",
						"Mountain Standard Time (Mexico)", "Myanmar Standard Time", "N. Central Asia Standard Time",
						"Namibia Standard Time", "Nepal Standard Time", "New Zealand Standard Time",
						"Newfoundland Standard Time", "Norfolk Standard Time", "North Asia East Standard Time",
						"North Asia Standard Time", "North Korea Standard Time", "Omsk Standard Time",
						"Pacific SA Standard Time", "Pacific Standard Time", "Pacific Standard Time (Mexico)",
						"Pakistan Standard Time", "Paraguay Standard Time", "Qyzylorda Standard Time",
						"Romance Standard Time", "Russia Time Zone 10", "Russia Time Zone 11",
						"Russia Time Zone 3", "Russian Standard Time", "SA Eastern Standard Time",
						"SA Pacific Standard Time", "SA Western Standard Time", "Saint Pierre Standard Time",
						"Sakhalin Standard Time", "Samoa Standard Time", "Sao Tome Standard Time",
						"Saratov Standard Time", "SE Asia Standard Time", "Singapore Standard Time",
						"South Africa Standard Time", "South Sudan Standard Time", "Sri Lanka Standard Time",
						"Sudan Standard Time", "Syria Standard Time", "Taipei Standard Time",
						"Tasmania Standard Time", "Tocantins Standard Time", "Tokyo Standard Time",
						"Tomsk Standard Time", "Tonga Standard Time", "Transbaikal Standard Time",
						"Turkey Standard Time", "Turks And Caicos Standard Time", "Ulaanbaatar Standard Time",
						"US Eastern Standard Time", "US Mountain Standard Time", "UTC",
						"UTC-02", "UTC-08", "UTC-09", "UTC-11", "UTC+12", "UTC+13",
						"Venezuela Standard Time", "Vladivostok Standard Time", "Volgograd Standard Time",
						"W. Australia Standard Time", "W. Central Africa Standard Time", "W. Europe Standard Time",
						"W. Mongolia Standard Time", "West Asia Standard Time", "West Bank Standard Time",
						"West Pacific Standard Time", "Yakutsk Standard Time", "Yukon Standard Time",
					},
					Message: "time_zone must be a valid Windows time zone name",
				},
				{
					// install_patches.reboot enum (PossibleValuesForRebootOptions).
					PropertyPath: "properties.installPatches.rebootSetting",
					AllowedValues: []string{
						"Always",
						"IfRequired",
						"Never",
					},
					Message: "reboot must be one of Always, IfRequired, Never",
				},
			},
			// NOTE: in_guest_user_patch_mode ("Platform"/"User") is mapped by AzureRM into
			// the free-form properties.extensionProperties map ("InGuestPatchMode" key), and
			// linux/windows classifications_to_include are array elements
			// (properties.installPatches.{linux,windows}Parameters.classificationsToInclude[*]);
			// neither is a single fixed ARM body path, so no declarative rule is emitted.
		},
	}
}

func init() { azwise.Register(NewMaintenanceConfiguration()) }
