# Azapin Generator — Technical Specification

This document describes the technical rules and implementation details of the azapin schema generator. It complements DESIGN.md (architecture and project scope) with the specific behaviors that the generator must enforce.

## Schema Flag Derivation

The generator reads bicep property flags (`flags` field in `types.json`) and derives Terraform schema flags. The mapping is not 1:1 — several rules apply.

### Flag Values

| Bit | Name | Value | Meaning |
|---|---|---|---|
| 0 | Required | 1 | Property must be set by the user |
| 1 | ReadOnly | 2 | Server-populated, not settable |
| 2 | WriteOnly | 4 | Accepted in PUT, never returned in GET |
| 3 | SystemManaged | 8 | Internal to ARM envelope (id, type, apiVersion) |

### Terraform Flag Rules

**Rule 1: Required properties → `Required: true`**

If the bicep flag has bit 0 set, the Terraform attribute is `Required: true`. This is the highest priority — Required overrides all other flags.

**Rule 2: ReadOnly properties → `Computed: true` (no Optional)**

If the bicep flag has bit 1 set, the Terraform attribute is `Computed: true` only. The user cannot set it; the server populates it. This applies to properties like `provisioningState`, `creationTime`, `primaryEndpoints`.

**Rule 3: Optional properties → `Optional: true, Computed: true` (safe default)**

If the property is neither Required nor ReadOnly, it is `Optional: true, Computed: true`. This is the safe default because most ARM properties appear in GET responses with server-populated values even when the user didn't explicitly set them. Marking as `Optional`-only would cause permanent diffs when the server returns a default value but Terraform expects null.

`Computed` should only be removed from `Optional` properties when there is explicit evidence that the server never populates the field unless the user sets it. This evidence comes from out-of-band sources (azwise knowledge, override files, manual testing), not from the bicep types alone.

When a known default value is available (from azwise or an override file), the property should use `Optional: true` with a `Default` plan modifier instead of `Optional + Computed`.

**Rule 4: WriteOnly properties → `Sensitive: true`**

If bit 2 is set, the property is accepted in PUT but never returned in GET. These map to `Sensitive: true` in Terraform.

**Rule 5: SystemManaged properties → skipped entirely**

Properties with bit 3 set (`id`, `type`, `apiVersion`, `name`) are ARM envelope fields managed by the provider framework. They are not emitted as schema attributes.

### Block-Level Computed Inference

Bicep flags are per-property. But Terraform needs block-level flags too. The generator applies these inference rules:

**Rule 6: Fully-computed blocks → `Computed: true`**

If an ObjectType (block) has ALL leaf descendants ReadOnly, the block itself is `Computed: true` even if its own bicep flag says Optional. The user never sets any part of this block.

Example: `primaryEndpoints` is `flags=2` (ReadOnly), and all its children (`blob`, `queue`, `table`, etc.) are also ReadOnly. The `ipv6Endpoints` sub-block has `flags=0` (Optional in bicep) but ALL its children are ReadOnly — so azapin emits it as `Computed: true`.

**Rule 7: Settable blocks → `Optional: true, Computed: true` (safe default)**

If an ObjectType has at least one settable (non-ReadOnly) leaf descendant, the block follows Rule 3: `Optional: true, Computed: true`. The server may return the block in GET with default values even when the user didn't set it.

Example: `sasPolicy` has `flags=0` and its children (`sasExpirationPeriod`, `expirationAction`) are Required. The block is `Optional: true, Computed: true` — the user can set it, and the server may return it.

## Validators

**Rule 8: No validators on Computed-only fields**

Computed fields are populated by the server. Validators constrain user input, which doesn't exist for computed fields. The generator never emits `Validators` for fields where the effective flag is Computed-only.

**Rule 9: Enum validators for settable enum fields**

String enum types (UnionType of StringLiteralType values) emit `stringvalidator.OneOf(...)` with all allowed values — but only when the field is settable (Required or Optional).

## Property Name Conversion

**Rule 10: camelCase → snake_case with stored ARM name**

The Terraform SDK requires `^[a-z_][a-z0-9_]*$` for attribute names. The generator converts ARM camelCase to Terraform snake_case using `naming.CamelToSnake()`.

The original ARM property name is NOT stored in a property map in the generated code. Instead, at runtime the provider reads the bicep type graph (already embedded in `internal/azure/generated/`) to reconstruct the ARM JSON payload. This avoids duplicating the ARM name data.

Algorithm:
- Insert `_` before an uppercase letter that follows a lowercase letter or digit
- Insert `_` at the end of an uppercase acronym (before the next lowercase): `isHNSEnabled` → `is_hns_enabled`
- `minimumTlsVersion` → `minimum_tls_version`
- `iPRules` → `ip_rules`
- `isNfsV3Enabled` → `is_nfs_v3_enabled`

**Rule 11: Build-time collision detection**

The generator checks all 3,246 ARM resource types for Terraform name collisions. 9 collisions exist (Admin namespaces, duplicate naming); these are resolved via an override table.

## Runtime Property Mapping

**Rule 12: No static property map — use embedded type graph**

The generated code does NOT include a `PropertyMap` variable. Instead, at runtime the CRUD methods use the bicep type graph already embedded in `internal/azure/generated/` to:

1. Walk the Terraform state attributes
2. Look up the corresponding ARM property name from the type graph
3. Build the ARM JSON payload with correct camelCase keys
4. On GET response, reverse-map ARM JSON keys back to Terraform attribute names

This avoids duplicating ~215 property mappings per resource (storage account has 215 nested paths) across 2,631 resource types.

## Type Mapping

| Bicep Type | Terraform Type | Schema Attribute |
|---|---|---|
| StringType | `types.StringType` | `schema.StringAttribute` |
| StringLiteralType | `types.StringType` | (enum value — used in UnionType) |
| IntegerType | `types.Int64Type` | `schema.Int64Attribute` |
| BooleanType | `types.BoolType` | `schema.BoolAttribute` |
| ObjectType | `types.ObjectType` | `schema.SingleNestedAttribute` |
| ArrayType(ObjectType) | `types.ListType` | `schema.ListNestedAttribute` |
| ArrayType(primitive) | `types.ListType` | `schema.ListAttribute` |
| UnionType(StringLiterals) | `types.StringType` | `schema.StringAttribute` + `stringvalidator.OneOf` |
| UnionType(mixed) | `types.StringType` | `schema.StringAttribute` (fallback) |
| AnyType | `types.DynamicType` | `schema.DynamicAttribute` |
| DiscriminatedObjectType | `types.DynamicType` | `schema.DynamicAttribute` (future: flattened variants) |

## Scope Selection

**Rule 13: Latest stable API version only**

For each ARM resource type, the generator selects the latest non-preview API version. This gives ~2,631 resources. Preview-only resource types are excluded — users fall back to `azapi_resource`.

**Rule 14: One Terraform resource per ARM resource type**

Each ARM resource type produces exactly one Terraform resource, pinned to one API version. The resource name encodes the service and resource path but NOT the API version.

## Generated Code Structure

```go
// Code generated by azapin; DO NOT EDIT.
package generated

import (
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
    "github.com/hashicorp/terraform-plugin-framework/schema/validator"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

// AzapiStorageAccountSchema returns the Terraform resource schema.
func AzapiStorageAccountSchema() schema.Schema {
    return schema.Schema{
        Description: "Manages a Microsoft.Storage/storageAccounts resource.",
        Attributes: map[string]schema.Attribute{
            "kind": schema.StringAttribute{
                Description: "Required. Indicates the type of storage account.",
                Required:    true,
                Validators: []validator.String{
                    stringvalidator.OneOf("Storage", "StorageV2", "BlobStorage", ...),
                },
            },
            "properties": schema.SingleNestedAttribute{
                Optional: true,
                Attributes: map[string]schema.Attribute{
                    "access_tier": schema.StringAttribute{
                        Optional: true,
                        Validators: []validator.String{
                            stringvalidator.OneOf("Hot", "Cool", "Premium", "Cold"),
                        },
                    },
                    "provisioning_state": schema.StringAttribute{
                        Computed: true,
                        // No validators — computed-only
                    },
                    ...
                },
            },
            "primary_endpoints": schema.SingleNestedAttribute{
                Computed: true,  // All children are ReadOnly
                Attributes: map[string]schema.Attribute{
                    "blob": schema.StringAttribute{
                        Computed: true,
                    },
                    ...
                },
            },
            ...
        },
    }
}
```

## Future Enhancements

### Default Values

When azwise knowledge (or another source) provides actual ARM default values, the generator can emit `Default: value` for Optional properties instead of bare `Optional: true`. This gives better plan-time behavior: Terraform shows the default in plan output rather than "(known after apply)".

### Discriminated Unions

Currently emitted as `DynamicAttribute`. Future work:
- Simple unions (2-3 variants): flatten all variant properties with `ExactlyOneOf` validators
- Complex unions: keep as `DynamicAttribute` with per-variant documentation

### Plan Modifiers

Type-specific plan modifiers (`UseStateForUnknown`, `RequiresReplace`) can be added per-type:
- `stringplanmodifier.UseStateForUnknown()` for Computed-only string attributes
- `stringplanmodifier.RequiresReplace()` for ForceNew properties (from azwise overlays)

These require type-matched imports (`stringplanmodifier`, `boolplanmodifier`, `int64planmodifier`, `objectplanmodifier`) which the generator must select based on attribute type.
