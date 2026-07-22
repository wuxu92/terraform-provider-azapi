# Azapin Generator — Technical Specification

This document describes the 17 technical rules and implementation details of the azapin schema generator. It complements DEVELOPER_SPEC.md (architecture, naming, project scope, and design decisions) with the specific behaviors that the generator must enforce.

## Adding a New Resource (end-to-end)

> For the full worked walkthrough (with the azwise extraction in detail) plus the
> **update/regenerate** and **API-version-upgrade** workflows, see
> [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md). The checklist below is the terse version.

Generating a new static resource from the ground up — with
`Microsoft.Storage/storageAccounts/blobServices` (→ `azapi_storage_account_blob_service`)
as the worked example — follows this ordered checklist. **Do not skip the azwise
step (2):** without it the generated schema carries only bicep-derived flags and
loses the curated AzureRM validation, defaults, and ForceNew knowledge.

1. **Register the ARM type.** Add the fully-qualified type constant to
   `internal/native/armtype/armtype.go` (e.g. `StorageAccountBlobService =
   "Microsoft.Storage/storageAccounts/blobServices"`).

2. **Learn AzureRM knowledge (azwise).** Extract operational knowledge — ForceNew,
   validation (enum/regex/length/range), defaults, computed/sensitive fields,
   timeouts — from `terraform-provider-azurerm` into a knowledge file
   `github.com/wuxu92/azwise/services/<service>/<resource>.go`, which self-registers
   via `func init() { azwise.Register(New…()) }` (a new service package also gets a
   blank import in the azwise repo's `services/all/all.go`). Use the **azwise
   agent/skill** (`.omp/agents/azwise.md`,
   `skill://azwise`), which drives the `azwise_extract` tool over the azurerm source.
   - **Sub-service API separation:** settings AzureRM bundles into a parent block
     (e.g. `blob_properties`, `share_properties` inside `azurerm_storage_account`)
     are *separate* ARM resources. Their knowledge goes in the sub-service's own
     file, never the parent's.
   - The generator overlays this automatically via `ApplyAzwise` (Rule 9b);
     declarative rules become baked schema validators/defaults/modifiers.

3. **Add a generation target.** Append the ARM type to the target list in
   `internal/native/generator/cmd/generate_poc.go`.

4. **Add a customizer (only if needed).** For rules neither bicep nor azwise can
   express — a singleton name pinned to `OneOf("default")`, a resource-specific
   semantic validator, a parent-name override — add
   `internal/native/generator/customizers/<resource>.go` and register it (Rule 9d).
   Skip when the bicep + azwise knowledge already suffices.

5. **Regenerate.** Run `go run ./internal/native/generator/cmd/generate_poc.go`.
   The generated `<name>_gen.go` self-registers via `init()`, so the provider
   exposes it through `services.Registry` with no provider edits.

6. **Validate.** `go run ./internal/native/cmd/azapin-validate/` must report
   `0 mismatches` for the new resource (schema ↔ bicep parity).

7. **Test.** Add a runtime composition test (`resource/base_test.go`), an
   acceptance `Describe` file (`acceptance/`), and an azwise rule test
   (`github.com/wuxu92/azwise/`).

8. **Document.** Refresh the DEVELOPER_SPEC.md "Current status" inventory (§7) if the
   resource set or component status changed.

## Schema Flag Derivation

The generator reads bicep property flags (`flags` field in `types.json`) and derives Terraform schema flags. The mapping is not 1:1 — several inference, post-processing, and description-mining rules apply.

### Flag Values

| Bit | Name | Value | Meaning |
|---|---|---|---|
| 0 | Required | 1 | Property must be set by the user |
| 1 | ReadOnly | 2 | Server-populated, not settable |
| 2 | WriteOnly | 4 | Accepted in PUT, never returned in GET |
| 3 | SystemManaged | 8 | Internal to ARM envelope (id, type, apiVersion) |

### Terraform Flag Rules

**Rule 1: Required properties → `Required: true`**

If the bicep flag has bit 0 set, the Terraform attribute is `Required: true` — unless a `Default` is available (Rule 8), which demotes the property to `Optional + Computed + Default` since a field with a curated default is omittable. Apart from that, `Required` overrides the other flag rules.

**Rule 2: ReadOnly properties → `Computed: true` (no Optional)**

If the bicep flag has bit 1 set, the Terraform attribute is `Computed: true` only. The user cannot set it; the server populates it. This applies to properties like `provisioningState`, `creationTime`, `primaryEndpoints`.

**Rule 3: Optional properties → `Optional: true, Computed: true` (safe default)**

If the property is neither Required nor ReadOnly, it is `Optional: true, Computed: true`. This is the safe default because most ARM properties appear in GET responses with server-populated values even when the user didn't explicitly set them. Marking as `Optional`-only would cause permanent diffs when the server returns a default value but Terraform expects null.

`Computed` should only be removed from `Optional` properties when there is explicit evidence that the server never populates the field unless the user sets it. This evidence comes from out-of-band sources (azwise knowledge, override files, manual testing), not from the bicep types alone.

When a known default value is available (from description extraction, azwise, or a customizer), the property is emitted as `Optional: true, Computed: true` with a `Default` (see Rule 8). The framework **requires** `Computed` whenever a `Default` is set — `Default` is a state value, not a plan modifier, and a non-`Computed` `Default` is rejected at provider startup. What the default *does* replace is the Rule 3 `UseStateForUnknown` safe-default modifier, which is omitted for default-bearing attributes.

**Rule 4: WriteOnly properties stay normal Optional+Computed attributes**

If bit 2 is set, the property is accepted in PUT but may be omitted from ordinary GET responses. Do **not** map this to Terraform `Sensitive`: framework sensitivity is a UI redaction/state-display concern, not a mutability or read-back rule. Secret redaction comes only from azwise `SensitiveFields` leaf paths. Runtime flattening preserves write-only values across read/apply when ARM omits or echoes them in an incompatible form.

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

## Post-Processing Rules

These rules are applied after parsing the type graph (`PostProcess()`), before schema emission. They derive information that bicep flags alone don't capture.

### Default Value Extraction

**Rule 8: Extract defaults from descriptions**

Many ARM property descriptions mention default values even when the OpenAPI/bicep spec doesn't formally declare them. The generator scans descriptions for patterns like:

- `"The default value is true"` → bool default `true`
- `"defaults to TLS1_2"` → string/enum default `"TLS1_2"`
- `"default is NoRootSquash"` → enum default `"NoRootSquash"`
- `"enabled by default"` → bool default `true`
- `"disabled by default"` → bool default `false`

When a valid default is extracted and type-checked against the property type:
- The property is emitted as `Optional: true, Computed: true` with a `Default`. The
  framework requires `Computed` whenever a `Default` is set; it rejects a `Default`
  on a non-`Computed` attribute at provider startup.
- A `Default` **demotes a `Required` property to Optional** (`demoteDefaultedRequired`):
  a field with a curated default is omittable, so it is never emitted `Required` — the
  framework forbids `Required + Default`. E.g. `kind` → `StorageV2` and network
  `default_action` → `Allow`, matching azurerm's optional-with-default fields.
- Implementations: the framework's own `booldefault.StaticBool`, `stringdefault.StaticString`, `int64default.StaticInt64` (emitted directly; enum defaults use `stringdefault.StaticString`)

Validation rules for extracted defaults:
- Bool defaults must be "true" or "false"
- Enum defaults must match one of the allowed values (case-insensitive)
- Int defaults must be numeric
- "null" and "undefined" are rejected

Example: `supportsHttpsTrafficOnly` description says "The default value is true since API version 2019-04-01" → emitted as `Optional: true, Computed: true, Default: booldefault.StaticBool(true)`.

### Schema Flag Invariants (generation-time guard)

`CheckFlagInvariants` (`invariants.go`) runs after post-processing and customizers,
before emission, and fails generation on framework attribute-flag rules the bicep
types and the framework's own startup validation don't catch:

- **A `Default` requires `Optional + Computed`** — never `Required` or
  read-only/computed-only. The framework rejects a non-`Computed` `Default` and
  forbids `Required + Default`, and the emitter would otherwise drop the default
  silently. (So a property with a default doesn't need the Computed *safe-default*
  of Rule 3, but it still needs the Computed *flag*.)
- **A `Default` must satisfy the attribute's own validators** — an enum/range/length
  default outside its validated set would make the provider reject the value it
  itself supplies when the field is omitted. The framework never cross-checks this.

`azapin-validate` reports the same violations as `INVARIANT …` lines. Because
customizers and azwise rules run *before* the guard, a contradictory hand-written
rule is caught at generation time rather than at provider startup.

### Single-Optional-Child Promotion

**Rule 9: Single optional child in a block → Required**

If an ObjectType has exactly one Optional property and zero Required properties (ignoring SystemManaged and ReadOnly properties), the Optional property is promoted to Required.

Rationale: a block with only one settable property serves no purpose when that property is absent — the user would be creating an empty block. Making it Required ensures the block is meaningful when present.

Example: `Placement` has one property `zonePlacementPolicy` (Optional in bicep). After promotion, `zone_placement_policy` becomes Required within the `placement` block. Other examples: `CorsRules.corsRules`, `Multichannel.enabled`, `DualStackEndpointPreference.publishIpv6Endpoint`.

### Azwise Knowledge Overlay

**Rule 9b: Overlay curated AzureRM knowledge (azwise)**

After the heuristic post-processing steps, `ApplyAzwise(def)` overlays the
hand-verified AzureRM-derived knowledge from `github.com/wuxu92/azwise` onto the type
graph. azwise rules are authoritative and take precedence over the bicep-flag and
description-mined heuristics. It is a no-op when no knowledge is registered for the
resource type. Property paths are ARM dot paths (`properties.accessTier`,
`sku.name`); paths that don't resolve in the body (e.g. ListKeys-only sensitive
fields) are skipped.

| azwise data | schema effect |
|---|---|
| `ForceNew` paths | `prop.ForceNew` → type-specific `RequiresReplace()` plan modifier (settable attributes only) |
| `ComputedFields` | `prop.ForceComputed` → forces `Computed: true` (overrides Rule 3 safe default) |
| `SensitiveFields` | `prop.Sensitive` → `Sensitive: true` |
| `DefaultValues` (non-nil) | `prop.DefaultValue` → `Default(...)` (overrides description-mined default) |
| `StringRules` (path-scoped) | `OneOf` (non-enum only), `LengthBetween/AtLeast/AtMost`, `RegexMatches` |
| `IntRules` | `int64validator.Between/AtLeast/AtMost` |

The declarative `ForceNew` list becomes schema-level `RequiresReplace`; **conditional**
ForceNew logic (e.g. storage SKU zone migration) is NOT expressible as a plan
modifier and stays in `azwise.CheckForceNew`, consulted by the resource's
`ModifyPlan` (see RESOURCE.md). The overlay only adds flags/modifiers/validators —
it never adds or removes properties, so the schema↔bicep validator still passes.

Because azwise is authoritative, an azwise validator **replaces** a same-kind
validator the description-miner produced for the same property (e.g. a
retention-day `IntBetween` that also appears in the ARM description) instead of
stacking a duplicate (`dropValidators` in `azwise_overlay.go`).

Storage account example: 10 `RequiresReplace` modifiers (`is_hns_enabled` →
`boolplanmodifier`, `sku.tier` → `stringplanmodifier`, `extended_location` →
`objectplanmodifier`), plus verified enum defaults (`minimum_tls_version` = `TLS1_2`,
`access_tier` = `Hot`).

### Operational Envelope (generated, not runtime)

**Rule 9c: Synthesize the operational envelope at generation time**

The bicep body graph carries `location`/`tags`/`identity`/`sku`/`properties`, but
not the attributes that map to the ARM resource **ID**: the resource `name`, the
parent reference, and the computed `id`. The generator synthesizes these three —
the *operational envelope* — directly into the generated schema, so the shipped
schema is complete and the runtime never mutates it (the only runtime addition is
the `timeouts` block, which needs a `context.Context` the static function cannot
hold).

`PostProcess` seeds `def.Envelope` from `naming.ParentReference(armType, writableScopes)`
— the single source of truth for the parent attribute name **and** its ID-shape
validator:

| Resource shape | Parent attribute | Validator |
|---|---|---|
| Resource-group scoped | `resource_group_id` | `/subscriptions/{}/resourceGroups/{}` |
| Subscription scoped | `subscription_id` | `/subscriptions/{}` |
| Management-group scoped | `management_group_id` | `/providers/Microsoft.Management/managementGroups/{}` |
| Child resource | `<parent-type>_id` (e.g. `storage_account_id`) | parent ID interleave pattern |
| Ambiguous (multi-scope/extension/tenant) | `parent_id` | none |

`name` and the parent reference are emitted `Required` + `RequiresReplace`; `id` is
`Computed` + `UseStateForUnknown`. The chosen parent name is also written to the
descriptor as `ParentAttr` so the runtime can compose/parse the ARM ID without
recomputing it. The synthesized envelope attributes are excluded from the
schema↔bicep validator via `EnvelopeAttrNames` (they have no bicep counterpart).

**Meta attributes (synthetic, behavior-only).** A customizer may append a
`typegraph.MetaAttr` to `def.Envelope.Meta` for a top-level attribute that is *not*
part of the bicep body and is never sent to or read from ARM — it drives
provider-side behavior only (e.g. `purge_on_destroy` on `azapi_key_vault`, read by
an `AfterDelete` hook to purge the soft-deleted shadow). The emitter renders each as
an `Optional`-only `BoolAttribute` (no default), so it stays null when unset and
never drifts on read/import; the mapper skips it on `Expand` and leaves it null on
flatten. Like `name`/parent/`id`, Meta attributes have no bicep counterpart and are
excluded from the schema↔bicep validator via `EnvelopeAttrNames`. Pair a Meta
attribute with the runtime hook that reads it — the schema flag and the behavior are
one feature; ship them together.

### Schema Customization Plugins

**Rule 9d: Per-resource customizers run last and bake into the generated file**

Some schema rules can't be inferred from bicep or azwise — e.g. the storage
account name charset (the name isn't even in the body graph). Developers express
these in Go as a **customizer** in the dedicated `customizers` sub-package under
the generator (`internal/native/generator/customizers`). Each resource gets its
own `<resource>.go` file holding just the customizer function, and a single
`Register` line in `register.go` (the one place that owns the `init()` and wires
every customizer, keyed by ARM type):

```go
// storage_account.go — the per-resource logic
func customizeStorageAccount(def *typegraph.ResourceDefinition) {
    // name is part of the operational envelope, not the bicep body
    def.SetNameValidators(
        typegraph.LengthValidator(3, 24),
        typegraph.RegexValidator(`^[a-z0-9]+$`, "must be 3-24 lowercase letters and digits"),
    )
    // mark body properties by ARM dot path; calls chain and each path is variadic
    // (the helpers panic on an unknown path, so a typo fails the generator)
    def.Default("properties.minimumTlsVersion", "TLS1_2").
        Required("sku", "sku.name")
}

// register.go — the single registration point (ARM types come from the armtype catalog)
func init() {
    Register(armtype.StorageAccount, customizeStorageAccount)
}
```

The customizer mechanism lives entirely in the sub-package; generator core does
**not** depend on it (so the runtime, which calls `generator.PostProcess` to load
body type graphs, never pulls customizers in). The generator command imports the
`customizers` package — which runs the `register.go` `init()` — and calls
`customizers.Apply(defs)` **after** `generator.PostProcess`, so a hand-written
customizer runs after the azwise overlay and envelope defaults and has the final
say. Because it mutates the same type graph and `Envelope` spec the emitter
consumes, every change is baked into the generated `_gen.go` — there is no runtime
schema mutation and no runtime cost.

Prefer the fluent helper methods on `*typegraph.ResourceDefinition` (in
`typegraph/customize.go`) — each is path-variadic, chains, and panics on an
unknown path, so a customizer reads as intent rather than field-poking:

| Helper | Effect |
|---|---|
| `Required(paths…)` | promote to `FlagRequired` |
| `Computed(paths…)` | force `ForceComputed` (Computed-only) |
| `ForceNew(paths…)` | `ForceNew` (RequiresReplace plan modifier) |
| `Sensitive(paths…)` | `Sensitive: true` |
| `AsSet(paths…)` | emit a primitive array as a set (`UseSet`) |
| `WithEmptyListDefault(paths…)` | empty-list schema default (`DefaultEmptyList`) |
| `NonNullStateForUnknown(paths…)` | opt out of `UseStateForUnknown` for null→non-null server fields |
| `Default(path, value)` | set `DefaultValue` |
| `AddValidatorsFor(path, validators…)` | append `Validators` |
| `SetNameValidators(validators…)` | replace the envelope name attribute's validators |
| `SetParent(name, description)` | rename/redescribe the envelope parent reference |
| `AddMetaAttr(attrs…)` | append synthetic behavior-only envelope attributes |

Underneath, these mutate the same `*Property` fields (`DefaultValue`, `ForceNew`,
`Sensitive`, `ForceComputed`, `Validators`, `Description`) and `Envelope` spec you
may still set directly for the rare case a helper does not cover; validators are
built with the `typegraph.RegexValidator` / `LengthValidator` / `OneOfValidator` /
`IntRangeValidator` constructors (which build the `DescriptionValidator` the
emitter already knows how to emit). `typegraph.FindProperty` and
`typegraph.IsolateArrayElement` (and every path-based helper above) **panic** on
an unresolved/invalid ARM path — a customizer authors paths by hand, so a typo
must fail generation rather than be silently skipped.

For a semantic rule a regex/length/enum cannot express, attach a hand-written
`validator.String` and reference its constructor by its real Go symbol through
`typegraph.Validator(<ctor>)` — e.g. `typegraph.Validator(validators.UUID)`, where
`validators.UUID` is the uncalled `func() validator.String`. Passing the symbol
(not a string) means a rename or deletion of the constructor is a compile error at
the customizer callsite. Validators live in two homes by scope, one per file:

- **Generic / cross-resource** — `internal/native/schema/validators` (package
  `validators`): `validators.UUID`, `validators.AzureResourceID` (ports of
  AzureRM's `validation.IsUUID` / `azure.ValidateResourceID`). Reuse these across
  resources instead of re-implementing per service.
- **Service-specific** — `internal/native/services/<service>/validators` (also
  package `validators`): `storagevalidators.StorageAccountIPRule`,
  `networkvalidators.VirtualNetworkBgpCommunity`, etc. — a rule tied to one
  service lives with that service, imported under a `<svc>validators` alias.

Add a new validator by writing just the `validator.String` constructor in the
package that matches its scope and referencing it from the customizer — no
registration, no catalog. The generator never imports either package: the emitter
reflects the passed func value (`runtime.FuncForPC`) to recover the import path,
then derives the call qualifier and import alias from it (bare `validators` for the
shared package, `<svc>validators` for a service-local one, so two packages that
share the base name `validators` never collide), and bakes them into the generated
schema, which is what actually links against the constructor.

When attaching a validator to one array's elements where sibling arrays share the
same deduplicated bicep element type (e.g. `ipRules` vs `ipv6Rules`), call
`generator.IsolateArrayElement(def, "<array path>")` first so the change does not
leak to the siblings.

`customizers.Register` panics on a duplicate registration — one customizer per
resource type.

## Validators

**Rule 10: No validators on Computed-only fields**

Computed fields are populated by the server. Validators constrain user input, which doesn't exist for computed fields. The generator never emits `Validators` for fields where the effective flag is Computed-only (including fields that are Computed via block-level inference in Rule 6).

**Rule 11: Enum validators for settable enum fields**

String enum types (UnionType of StringLiteralType values) emit `stringvalidator.OneOf(...)` with all allowed values — but only when the field is settable (Required or Optional).

**Rule 12: Description-based validators for settable fields**

When bicep types don't include formal validators, the generator extracts validation hints from property descriptions. Three patterns are recognized:

1. **ARM resource ID format**: Descriptions containing `/subscriptions/{subscriptionId}/resourceGroups/...` emit a regex validator matching the ARM resource ID pattern.
   - Example: `VirtualNetworkRule.id` → `stringvalidator.RegexMatches(regexp.MustCompile('^/subscriptions/[^/]+/resourceGroups/[^/]+/providers/'))`

2. **Datetime format**: Descriptions mentioning `datetime format`, `ISO 8601`, or `yyyy-MM-dd` patterns emit a regex validator for the datetime format.
   - Example: `minCreationTime` with "format 'yyyy-MM-ddTHH:mm:ssZ'" → date regex

3. **Numeric ranges**: Descriptions containing `must be greater than N`, `less than or equal to N`, `between N and M`, `minimum N`, `maximum N` emit `int64validator.AtLeast`, `AtMost`, or `Between`.
   - Example: `shareQuota` with "Must be greater than 0, and less than or equal to 5120" → `int64validator.Between(0, 5120)`

These validators are only emitted for settable (non-Computed) fields. The extraction is conservative — ambiguous descriptions are skipped.

## Property Name Conversion

**Rule 13: camelCase → snake_case**

The Terraform SDK requires `^[a-z_][a-z0-9_]*$` for attribute names — this is enforced in `fwschema/attribute_name_validation.go` with a hard error. The generator converts ARM camelCase to Terraform snake_case using `naming.CamelToSnake()`.

The original ARM property name is NOT stored in the generated code. At runtime, the provider looks up the ARM name from the embedded bicep type graph in `internal/azure/generated/`. This is a direct lookup per attribute, not a heuristic runtime conversion.

Algorithm:
- Insert `_` before an uppercase letter that follows a lowercase letter or digit
- Insert `_` at the end of an uppercase acronym (before the next lowercase)
- `minimumTlsVersion` → `minimum_tls_version`
- `isHnsEnabled` → `is_hns_enabled`
- `isNfsV3Enabled` → `is_nfs_v3_enabled`
- `iPRules` → `ip_rules`
- `AzureAD` → `azure_ad`

**Rule 14: Build-time collision detection**

The generator checks all 3,246 ARM resource types for Terraform name collisions. 9 collisions exist (Admin namespaces, duplicate naming); these are resolved via an override table. Collision test runs as `TestResourceNameNoCollisions` in the naming package.

## Runtime Property Mapping

**Rule 15: No static property map — use embedded type graph**

The generated code does NOT include a `PropertyMap` variable. Instead, at runtime the CRUD methods use the bicep type graph already embedded in `internal/azure/generated/` to:

1. Walk the Terraform state attributes
2. Look up the corresponding ARM property name from the type graph
3. Build the ARM JSON payload with correct camelCase keys
4. On GET response, reverse-map ARM JSON keys back to Terraform attribute names

This avoids duplicating ~215 property mappings per resource (storage account has 215 nested paths) across 2,631 resource types. The type graph is already in the binary (334 MB) — reusing it costs zero additional binary size.

## Type Mapping

| Bicep Type | Terraform Type | Schema Attribute |
|---|---|---|
| StringType | `types.StringType` | `schema.StringAttribute` |
| StringLiteralType | (enum value) | Used in UnionType for enum validators |
| IntegerType | `types.Int64Type` | `schema.Int64Attribute` |
| BooleanType | `types.BoolType` | `schema.BoolAttribute` |
| ObjectType | `types.ObjectType` | `schema.SingleNestedAttribute` |
| ArrayType(ObjectType) | `types.ListType` | `schema.ListNestedAttribute` |
| ArrayType(primitive) | `types.ListType` | `schema.ListAttribute` with `ElementType` (`types.StringType`/`Int64Type`/`BoolType` per element kind) |
| UnionType(StringLiterals) | `types.StringType` | `schema.StringAttribute` + `stringvalidator.OneOf` |
| UnionType(mixed) | `types.StringType` | `schema.StringAttribute` (fallback) |
| AnyType | `types.DynamicType` | `schema.DynamicAttribute` |
| DiscriminatedObjectType | `types.DynamicType` | `schema.DynamicAttribute` (future: flattened) |

## Recursion and Discriminated Types

**Reference cycles.** Many ARM types are self-referential — the canonical case is
the `ErrorEntity`/`ErrorDetail` pattern (`details: ErrorDetail[]`). The walker
resolves `$ref` indices into a shared-pointer graph; to keep that graph a DAG it
tracks in-progress indices and replaces any back-edge (a reference to a node still
being built) with a `KindAny` sentinel. The recursive sub-tree therefore degrades
to a dynamic/string attribute instead of expanding forever. Without this, every
graph consumer (emitter, post-processing, validators) would recurse infinitely and
stack-overflow. Covered by `TestParseSelfReferentialType`.

**Discriminated objects.** A `DiscriminatedObjectType` stores its variants in an
`elements` field shaped as a JSON **object** (discriminator value → type), unlike
`UnionType` whose `elements` is an **array**. The parser keeps `elements` as
`json.RawMessage` and decodes it lazily per `$type`, so a discriminated type no
longer fails the whole document's unmarshal (previously ~28% of `types.json` files
errored out entirely). Discriminated bodies currently emit a `DynamicAttribute`;
a root body that resolves to a discriminated/non-object type is skipped (the
resource falls back to `azapi_resource`). Covered by `TestParseDiscriminatedObjectType`.

**Dynamic inside collections.** The framework forbids a dynamic type nested inside
a collection (`ListNestedAttribute`). A `KindAny` property (recursion sentinel or
discriminated/unknown type) therefore emits a `DynamicAttribute` only at the top
level or inside a `SingleNestedAttribute`; inside a `ListNested` element it degrades
to a `StringAttribute` instead. The generated schema is checked by
`schema.Schema.ValidateImplementation` in the resource layer's tests.

## Scope Selection

**Rule 16: Latest stable API version only**

For each ARM resource type, the generator selects the latest non-preview API version. This gives ~2,631 resources. Preview-only resource types are excluded — users fall back to `azapi_resource`.

**Implementation:** the bicep-types manifest `internal/azure/generated/index.json` is the single source of truth, wrapped by `generator.Index` (`LoadIndex`, `LatestStableVersion`, `TypesPath`, `ResolveLatestStable`). The manifest maps every `"<ARMType>@<version>"` to the types.json that defines it, so resolution reads it instead of scanning directories: `LatestStableVersion(armType)` takes the max over the type's non-preview versions (`YYYY-MM-DD` sorts chronologically; the `@` anchor stops a child type like `storageAccounts/blobServices` from leaking into the parent's lookup), and `TypesPath(tag)` returns the file straight from the entry's `$ref` — no `<service>/<namespace>/<version>` path is reconstructed. The generation command, the post-compile validator, and **every test that loads a bicep fixture** resolve through this one type. No version literal or namespace directory is hardcoded in the generation path, so vendoring a newer `index.json`/`types.json` automatically rolls the generated schema (and its tests) forward. A hardcoded version would silently pin generation to a stale API — exactly the failure this rule exists to prevent.

**Rule 17: One Terraform resource per ARM resource type**

Each ARM resource type produces exactly one Terraform resource, pinned to one API version. The resource name encodes the service and resource path but NOT the API version.

## Generated Code Structure

Generated resources are grouped by **service**. Each resource is written to
`internal/native/services/<service>/<name>_gen.go` (package `<service>`), where
`<service>` is the ARM namespace's service segment (`naming.ServiceName`, e.g.
`storage`) and `<name>` is the Terraform resource name without the `azapi_` prefix
(e.g. `services/storage/storage_account_gen.go`). The `_gen.go` suffix — together
with the in-file `// Code generated by azapin; DO NOT EDIT.` header — marks the
file as machine-generated. The path convention is centralized in
`generator.FileName`. Each service package self-registers into `services.Registry`
via `services.Register` in `init()`; the `services/all` package blank-imports
every service package so importing it populates the registry (add a line there per
new service). Generic hand-written validators live in
`internal/native/schema/validators/`; service-specific ones live in
`internal/native/services/<service>/validators/`.

Example output for `azapi_storage_account` (simplified):

```go
// Code generated by azapin; DO NOT EDIT.
package storage

import (
    "regexp"

    "github.com/Azure/terraform-provider-azapi/internal/native/services"
    "github.com/Azure/terraform-provider-azapi/internal/native/schema/validators"
    storagevalidators "github.com/Azure/terraform-provider-azapi/internal/native/services/storage/validators"
    nativeschema "github.com/Azure/terraform-provider-azapi/internal/native/schema"
    "github.com/Azure/terraform-provider-azapi/internal/native/schema/planmodifiers"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
    "github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func AzapiStorageAccountSchema() schema.Schema {
    return schema.Schema{
        Description: "Manages a Microsoft.Storage/storageAccounts resource.",
        Attributes: map[string]schema.Attribute{
            "kind": schema.StringAttribute{
                Required: true,
                Validators: []validator.String{
                    stringvalidator.OneOf("Storage", "StorageV2", "BlobStorage", ...),
                },
            },
            "properties": schema.SingleNestedAttribute{
                Optional: true,
                Computed: true,
                Attributes: map[string]schema.Attribute{
                    "access_tier": schema.StringAttribute{
                        Optional: true,
                        Computed: true,
                        Validators: []validator.String{
                            stringvalidator.OneOf("Hot", "Cool", "Premium", "Cold"),
                        },
                    },
                    "supports_https_traffic_only": schema.BoolAttribute{
                        Optional: true,
                        Default:  booldefault.StaticBool(true), // from description
                    },
                    "provisioning_state": schema.StringAttribute{
                        Computed: true, // ReadOnly — no validators
                    },
                    "network_acls": schema.SingleNestedAttribute{
                        Optional: true,
                        Computed: true,
                        Attributes: map[string]schema.Attribute{
                            "virtual_network_rules": schema.ListNestedAttribute{
                                ...
                                // id gets ARM resource ID regex from description
                            },
                        },
                    },
                },
            },
            "primary_endpoints": schema.SingleNestedAttribute{
                Computed: true, // All children ReadOnly — Rule 6
                Attributes: map[string]schema.Attribute{
                    "blob": schema.StringAttribute{Computed: true},
                    "ipv6_endpoints": schema.SingleNestedAttribute{
                        Computed: true, // All children ReadOnly — Rule 6
                        ...
                    },
                },
            },
        },
    }
}
```

The generated `Attributes` map also contains the synthesized operational envelope
(Rule 9c), and the file registers the resource via `init()`:

```go
        // operational envelope — name + parent reference + id
        "name": schema.StringAttribute{
            Required:      true,
            PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
            Validators: []validator.String{ // from a customizer (Rule 9d)
                stringvalidator.LengthBetween(3, 24),
                stringvalidator.RegexMatches(regexp.MustCompile(`^[a-z0-9]+$`), "..."),
            },
        },
        "resource_group_id": schema.StringAttribute{
            Required:      true,
            PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
            Validators: []validator.String{
                stringvalidator.RegexMatches(regexp.MustCompile(`(?i)^/subscriptions/[^/]+/resourceGroups/[^/]+$`), "..."),
            },
        },
        "id": schema.StringAttribute{
            Computed:      true,
            PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
        },
// ...
func init() {
    services.Register(services.Descriptor{
        Name:           "azapi_storage_account",
        ARMType:        "Microsoft.Storage/storageAccounts",
        APIVersion:     "2025-01-01",
        Schema:         AzapiStorageAccountSchema,
        WritableScopes: 8,
        ParentAttr:     "resource_group_id",
    })
}
```

### Conditional Imports

The emitter pre-scans the type graph and only includes imports that are actually used:
- `regexp`: only when description-based regex validators are emitted
- `int64validator`: only when numeric range validators are emitted
- `stringvalidator`: only when enum or regex validators are emitted
- `types`: only when `schema.ListAttribute` with `ElementType` is used
- `nativeschema`: only when a managed-identity attribute is emitted (`ManagedServiceIdentity`)
- `planmodifiers`: only when a location or discriminated-variant modifier (or a generic custom plan modifier) is emitted; a service-local `<svc>planmodifiers` is aliased

## Future Enhancements

### Discriminated Unions

Currently emitted as `DynamicAttribute`. Future work:
- Simple unions (2-3 variants): flatten all variant properties with `ExactlyOneOf` validators
- Complex unions: keep as `DynamicAttribute` with per-variant documentation

### Plan Modifiers

Type-specific plan modifiers (`UseStateForUnknown`, `RequiresReplace`) can be added per-type:
- `stringplanmodifier.UseStateForUnknown()` for Computed-only string attributes
- `stringplanmodifier.RequiresReplace()` for ForceNew properties (from azwise overlays)

These require type-matched imports (`stringplanmodifier`, `boolplanmodifier`, `int64planmodifier`, `objectplanmodifier`) which the generator must select based on attribute type.

### Override System

Per-resource overrides are now implemented by **customizer plugins** (Rule 9d):
strip `Computed` from user-only properties, add explicit defaults, mark ForceNew,
or correct ReadOnly inaccuracies — all by mutating the type graph in Go, baked
into the generated file. Possible future work: a declarative (non-Go) override
format for simple cases, and a lint that flags customizers whose target ARM path
no longer resolves after a bicep schema bump.
