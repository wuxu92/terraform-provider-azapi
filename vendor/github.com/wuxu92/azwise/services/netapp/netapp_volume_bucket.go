package netapp

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetAppVolumeBucket provides resource knowledge for
// Microsoft.NetApp/netAppAccounts/capacityPools/volumes/buckets.
//
// One ARM type shared by two typed Terraform resources — azurerm_netapp_volume_bucket
// and azurerm_netapp_volume_bucket_with_server — which share a common argument set;
// the _with_server variant additionally sets the properties.server sub-object. Only
// knowledge universal to both is unioned as top-level rules; value constraints that fire
// only when the server sub-object is present are safe to include (they never trigger for
// the plain bucket).
//
// Sources:
//   - internal/services/netapp/netapp_volume_bucket_resource.go
//     (Create 53-102 → Bucket{Properties{Path, Permissions, FileSystemUser{NfsUser, CifsUser}}};
//     timeouts Create/Update/Delete 60m Read 5m).
//   - internal/services/netapp/netapp_volume_bucket_with_server_resource.go
//     (Arguments 46-83 add server block: fqdn, certificate_pem, on_certificate_conflict_action
//     default Fail; key_vault ConflictsWith server.0.certificate_pem).
//   - internal/services/netapp/netapp_volume_bucket_helper.go
//     (common args 20-108: name ForceNew+BucketName; volume_id parent ForceNew;
//     file_system_nfs_user{group_id/user_id IntAtLeast(0)} ExactlyOneOf file_system_cifs_username;
//     key_vault block; path Optional ForceNew default "/" BucketPath; permissions default ReadOnly).
//   - internal/services/netapp/validate/bucket_name.go (3-63 chars, lowercase alnum/hyphen/period,
//     start/end alnum, no consecutive periods / ".-" / "-." sequences, not IPv4-shaped).
//   - internal/services/netapp/validate/bucket_path.go (must start with "/", no backslashes).
//   - go-azure-sdk resource-manager/netapp/2026-01-01/buckets:
//     model_bucketproperties.go (path/permissions/fileSystemUser/akvDetails/server settable;
//     provisioningState/status read-only), model_filesystemuser.go, model_nfsuser.go (groupId/userId),
//     model_bucketserverproperties.go (onCertificateConflictAction settable; certificateCommonName/
//     certificateExpiryDate/ipAddress read-only), constants.go (BucketPermissions ReadOnly/ReadWrite,
//     OnCertificateConflictAction Fail/Update), id_bucket.go (type segment casing "buckets").
type NetAppVolumeBucket struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetAppVolumeBucket)(nil)

// NewNetAppVolumeBucket returns knowledge for the buckets resource.
func NewNetAppVolumeBucket() *NetAppVolumeBucket {
	return &NetAppVolumeBucket{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.NetApp/netAppAccounts/capacityPools/volumes/buckets",
			ApiVersions:  []string{"2026-01-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.path"},
			},
			StringRules: []azwise.StringRule{
				{
					// AzureRM BucketName: 3-63 chars, lowercase letters/numbers/hyphens/periods,
					// must start and end with a lowercase letter or number.
					Regex:     `^[a-z0-9]([a-z0-9.-]*[a-z0-9])?$`,
					MinLength: 3,
					MaxLength: 63,
					Message:   "must be 3-63 characters, lowercase letters/numbers/hyphens/periods only, and start/end with a lowercase letter or number",
				},
				{
					// BucketPath: absolute POSIX-style path, no backslashes.
					PropertyPath: "properties.path",
					Regex:        `^/[^\\]*$`,
					Message:      "must be an absolute POSIX-style path starting with '/' and must not contain backslashes",
				},
				{
					PropertyPath:  "properties.permissions",
					AllowedValues: []string{"ReadOnly", "ReadWrite"},
				},
				{
					// Present only when the _with_server variant sets the server block.
					PropertyPath:  "properties.server.onCertificateConflictAction",
					AllowedValues: []string{"Fail", "Update"},
				},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.fileSystemUser.nfsUser.groupId", MinValue: azwise.Ptr(int64(0))},
				{PropertyPath: "properties.fileSystemUser.nfsUser.userId", MinValue: azwise.Ptr(int64(0))},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.path", Value: "/"},
				{PropertyPath: "properties.permissions", Value: "ReadOnly"},
			},
			// Exactly one file-system user identity must be supplied (both variants).
			ExactlyOneOf: []azwise.RelationalRule{
				{Paths: []string{"properties.fileSystemUser.nfsUser", "properties.fileSystemUser.cifsUser"}},
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.status",
				"properties.server.certificateCommonName",
				"properties.server.certificateExpiryDate",
				"properties.server.ipAddress",
			},
			// NOTE: the BucketName validator also forbids consecutive periods, ".-"/"-."
			// sequences, and IPv4-shaped names; RE2 (Go regexp) has no negative lookahead,
			// so those extra negative checks cannot be folded into the Regex above.
			// NOTE: _with_server's key_vault ConflictsWith server.0.certificate_pem
			// (properties.akvDetails vs properties.server.certificateObject) applies only when
			// the server block is present and is left to the resource customizer rather than a
			// declarative rule here.
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewNetAppVolumeBucket()) }
