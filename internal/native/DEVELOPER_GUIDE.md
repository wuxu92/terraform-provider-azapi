# Azapin Developer Guide

Task-oriented playbook for azapin static resources. Three workflows:

1. [Add a new resource from the ground up](#1-add-a-new-resource-from-the-ground-up) — incl. azwise knowledge extraction.
2. [Update / regenerate an existing resource](#2-update--regenerate-an-existing-resource) — bicep/azwise/customizer changed, same API version.
3. [Upgrade the API version of an existing resource](#3-upgrade-the-api-version-of-an-existing-resource).

Reference docs cover the *why*; this covers the *how*:
- **DEVELOPER_SPEC.md** — architecture, naming, scope, user requirements, design decisions.
- **GENERATOR.md** — the 17 generator rules (flags, defaults, validators, envelope, customizers).
- **RESOURCE.md** — the runtime `Base`, mapper, hooks.

## Mental model

A resource is assembled at **generation time** from four layers, each overriding the prior:

```
bicep flags (types.json)    Required/ReadOnly/WriteOnly → schema flags
 └ description mining        defaults & validators scraped from ARM descriptions
   └ azwise overlay          curated AzureRM knowledge (authoritative; Rule 9b)
     └ customizers           hand-written Go, runs last, final say (Rule 9d)
```

The result is baked into `internal/native/services/<service>/<name>_gen.go`. The
**runtime never mutates the schema** — it serves the generated `Schema()` plus a
`timeouts` block. Per-resource runtime *behavior* (not schema) lives beside the
generated resource in `<name>_hooks.go`; see [Runtime behavior hooks](#runtime-behavior-hooks).

Every resource also gets a **read-only data source** of the same name for free — no
per-resource work. The runtime converts the generated resource schema into a
data-source schema (`name` + parent Required, all else Computed) and reads via GET;
see RESOURCE.md "Data Source".

**Customizer vs hook — know which knob you're turning.** These are the two per-resource
extension points and they never overlap:
- **Customizer = generation-time, shapes the *schema*.** It runs inside the generator and
  mutates the intermediate `typegraph.ResourceDefinition` (the parsed type graph + envelope)
  *before* emission. Whatever it changes is baked into `<name>_gen.go` and frozen — flags
  (Required/Computed/ForceNew/Sensitive), validators, defaults, collection kind, envelope
  name/parent, meta attributes. If it alters what the attribute *looks like* in the
  Terraform schema, it's a customizer.
- **Hook = runtime, shapes *behavior*.** It runs in the provider at CRUD/plan time and never
  touches the schema — body massaging before PUT, read-side normalization, conditional
  `RequiresReplace`, cross-property validation, singleton/override wiring. If it changes what
  the resource *does* with a value (not how the value is declared), it's a hook.

A schema concern in a hook, or a behavior in a customizer, is always the wrong layer.

Two failure modes to internalize:
- **Customizer paths panic.** `FindProperty` / `IsolateArrayElement` crash generation
  on an unresolvable ARM path — a typo or renamed property can't ship silently.
- **azwise paths fail silently.** `ApplyAzwise` skips an unresolved path; a *lost* rule
  is caught only by the overlay tests (`TestApplyAzwise*`), which parse live
  `types.json` and assert the rule baked in.

## Where things live

> Task-oriented file map. For package **responsibilities** (what each package is
> for, rather than where to edit), see DEVELOPER_SPEC.md §3.3.

| What | Path |
|---|---|
| ARM type constants | `internal/native/armtype/armtype.go` |
| azwise knowledge (curated AzureRM) | `internal/azure/azwise/<resource>.go` + `register.go` |
| Generation targets | `internal/native/generator/cmd/generate_poc.go` |
| Customizers (generation-time schema) | `internal/native/generator/customizers/<resource>.go` + `register.go` |
| Schema validators (generic) | `internal/native/schema/validators/<rule>.go` |
| Schema validators (service-specific) | `internal/native/services/<service>/validators/<rule>.go` |
| Generated output | `internal/native/services/<service>/<name>_gen.go` |
| Runtime hooks for a generated resource | `internal/native/services/<service>/<name>_hooks.go` |
| Service aggregator (blank imports) | `internal/native/services/all/all.go` |
| Runtime base / mapper / hook registry | `internal/native/resource/` |
| Bicep types manifest | `internal/azure/generated/index.json` (+ `…/<ns>/<date>/types.json`) |
| Validator CLI | `internal/native/cmd/azapin-validate/` |

## The verification gate

Every workflow ends here (repo root):

```bash
# 1. format + compile + vet
gofmt -l internal/native/ internal/azure/azwise/      # must print nothing
go build ./...
go vet ./internal/native/... ./internal/azure/azwise/...

# 2. regenerate (schema is always generated, never hand-edited)
go run ./internal/native/generator/cmd/generate_poc.go

# 3. schema ↔ bicep parity — must report "0 mismatches" (and no INVARIANT lines)
go run ./internal/native/cmd/azapin-validate/

# 4. unit suites (offline)
go test ./internal/native/... ./internal/azure/azwise/...

# 5. acceptance (live Azure; needs TF_ACC=1 + ARM_SUBSCRIPTION_ID, else skips)
# **DO NOT** run acceptance tests with AI Agent
TF_ACC=1 go test ./internal/native/services/<service>/ -run <TestXxxAcceptance>
```

`azapin-validate` is load-bearing: it confirms the schema covers every bicep body
property and vice versa. `INVARIANT …` lines mean a `Default` violates the framework's
Optional+Computed/validator rules (GENERATOR.md "Schema Flag Invariants") — fix before
shipping.

---

## Runtime behavior hooks

Customizers shape the **schema** at generation time; hooks shape runtime **behavior**
and touch the schema not at all. Never validate/default/`RequiresReplace` a fixed
attribute from a hook (bake it into the schema); never read ARM state from a customizer.

Register per Terraform name from the generated service package. Hook files are
hand-written, named `<name>_hooks.go`, and live beside `<name>_gen.go`; the existing
`services/all` blank import for that service makes them compile into the provider:

```go
// internal/native/services/<service>/<name>_hooks.go
package <service>

func init() {
    resource.RegisterHooks(<Descriptor>.Name, &resource.Hooks{BeforeCreate: ..., ModifyPlan: ...})
}
```

`newBase` looks the bundle up once (`hookRegistry[name]`); a nil bundle — or a nil
field — means "use the base behavior".

### What you can customize

| Hook | Fires | Use it to |
|---|---|---|
| `BeforeCreate` | create, after body composed, before PUT | mutate `ctx.Body` before send |
| `AfterCreate` | create, after the GET following PUT | inspect `ctx.Response`; raise diagnostics |
| `BeforeUpdate` | update, after body composed, before PUT | mutate `ctx.Body` |
| `AfterUpdate` | update, after the follow-up GET | inspect `ctx.Response` |
| `BeforeRead` | read, before the GET | preflight/guard using `ctx.State` (e.g. skip or short-circuit) |
| `AfterRead` | read, after GET, before mapping into state | massage `ctx.Response` before flatten |
| `BeforeDelete` | delete, before the delete operation | preflight/guard using `ctx.State` |
| `Delete` | delete, instead of the default ARM DELETE | route destroy through a service-specific endpoint when ARM has no DELETE |
| `AfterDelete` | delete, after the delete operation succeeds | destroy-time orchestration once the resource is gone (e.g. purge a Key Vault's soft-deleted shadow) using `ctx.State` |
| `Relational` (data) | config validation (plan-time) | declarative cross-property constraints (`ExactlyOneOf` / `RequiredWith` / `ConflictsWith` / `AtLeastOneOf` / `AtMostOneOf`) over dot-path lists → framework `ConfigValidators` (see [Cross-property validation](#cross-property-validation)) |
| `ConfigValidators` (data) | config validation (plan-time) | plug prebuilt/reusable `resource.ConfigValidator` objects (e.g. from `terraform-plugin-framework-validators`) |
| `ValidateConfig` | config validation (plan-time; values may be unknown) | imperative cross-field rules not worth expressing as a validator object |
| `ModifyPlan` | plan, **after** the base's default work | conditional `RequiresReplace`, plan-time derivation |
| `Singleton` (data) | create/delete of a fixed-named default child ARM never creates or deletes | skip the create existence check + reset to `DefaultBody` via PUT on destroy (e.g. `blobServices/default`); the reset path runs none of the delete hooks |

`Before/AfterCreate` vs `Before/AfterUpdate` dispatch by whether the op is a create
(same body-composition path). `ValidateConfig`/`ModifyPlan` use framework signatures and
run *after* the base (base applies schema `RequiresReplace` first, then your `ModifyPlan`).

> `Before/AfterRead` bracket the GET: `BeforeRead` sees `ctx.State` (pre-GET), `AfterRead`
> sees `ctx.Response` (post-GET, before flatten).

### What a hook sees — `CrudCtx`

| Field | What | Notes |
|---|---|---|
| `Ctx` | operation `context.Context` | already timeout-scoped |
| `Client` | `*clients.Client` | make extra ARM calls if needed |
| `ID` | `parse.ResourceId` | the resource's Azure ID |
| `Plan` | typed plan object | live on create/update; **null** on read/delete |
| `State` | typed prior-state object | live on read/delete; **null** on create/update (the create/update path composes `Body` from `Plan`) |
| `Body` | `map[string]interface{}` — ARM body being composed | live on create/update; a mutation reaches ARM **only from a `Before*`** hook (the PUT reads it right after) |
| `Response` | `map[string]interface{}` — ARM GET response | live in `After*` (post-GET/read); **null** in every `Before*`; it is what gets flattened into state |
| `Diags` | `*diag.Diagnostics` | append an error + `return` to abort |

`Before*` mutate `Body`; `After*` massage `Response` (state is flattened from it). The base
checks `HasError()` after each hook — append a diagnostic and `return` to stop the operation.
The exact per-hook-point field-liveness matrix is the doc comment on `Hooks`/`CrudCtx` in
`internal/native/resource/hooks.go` (the authority, pinned to `Base` by
`base_hook_contract_test.go`); this table summarizes it.

### Example — conditional ForceNew (shipped)

A static plan modifier can't express "replace only when migrating between zonal and
non-zonal SKUs", so it lives in a `ModifyPlan` hook consulting `azwise.CheckForceNew`:

```go
// internal/native/services/storage/storage_account_hooks.go
func storageAccountModifyPlan(ctx context.Context, req fwresource.ModifyPlanRequest, resp *fwresource.ModifyPlanResponse) {
    if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
        return // updates only
    }
    skuName := path.Root("sku").AtName("name")
    var oldName, newName types.String
    resp.Diagnostics.Append(req.State.GetAttribute(ctx, skuName, &oldName)...)
    resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, skuName, &newName)...)
    if resp.Diagnostics.HasError() {
        return
    }
    // ARMType/APIVersion from the local descriptor — stays correct across version upgrades.
    oldBody := map[string]interface{}{"sku": map[string]interface{}{"name": oldName.ValueString()}}
    newBody := map[string]interface{}{"sku": map[string]interface{}{"name": newName.ValueString()}}
    if azwise.CheckForceNew(StorageAccount.ARMType, StorageAccount.APIVersion, oldBody, newBody) {
        resp.RequiresReplace = append(resp.RequiresReplace, path.Root("sku"))
    }
}
```

### Example — body / response massaging (template)

```go
resource.RegisterHooks(ExampleResource.Name, &resource.Hooks{
    BeforeCreate: func(c *resource.CrudCtx) { c.Body["someField"] = derive(c) },
    AfterRead:    func(c *resource.CrudCtx) { /* massage c.Response before flatten, or c.Diags.AddError(…) */ },
})
```

### Hook rules

- Schema concerns (validators, defaults, ForceNew on a fixed attribute) → generated
  schema via customizers/azwise, **never** a hook.
- Keyed by the **Terraform** name; register from `<name>_hooks.go` in the generated
  service package using `resource.RegisterHooks(<Descriptor>.Name, ...)`.
- Read ARM type/version from the local generated descriptor (`<Descriptor>.ARMType`,
  `<Descriptor>.APIVersion`), never hardcode.
- Append to `Diags`/`resp.Diagnostics` and `return` to abort; the base checks after each hook.
- If a hook strips or normalizes a server value on read (e.g. Azure's reserved-priority
  default access rule), the schema must **forbid a user from configuring that value** via a
  customizer validator/cap — otherwise every plan that sets it drifts. The read-side filter
  and the schema cap are one feature; ship them together.

### Cross-property validation

Rules that span **multiple** attributes (single-attribute validators — enum, range,
length — belong in the generated schema, not here). Three tiers, cheapest first:

**1. `Relational` (declarative, preferred).** A slice of `services.RelationalConstraint`,
each a kind plus a list of **dot-separated snake_case attribute paths** absolute from the
schema root. Lowered to framework `ConfigValidators` at runtime. Kinds: `ExactlyOneOf`,
`AtMostOneOf`, `AtLeastOneOf`, `RequiredWith` / `ConflictsWith` (path[0] is the subject).

```go
resource.RegisterHooks(VirtualNetwork.Name, &resource.Hooks{
    Relational: []services.RelationalConstraint{
        {Kind: services.ExactlyOneOf, Paths: []string{
            "properties.address_space.address_prefixes",
            "properties.address_space.ipam_pool_prefix_allocations",
        }, Message: "exactly one of address_space or ip_address_pool must be set"},
    },
})
```

> **Discriminated bodies must declare their own variant constraint here** — generation no
> longer synthesizes it. A required discriminated block → `ExactlyOneOf` over the variant
> paths; an optional one → `AtMostOneOf`.

Paths are plain strings the compiler can't check against the schema. A typo would surface
only when a practitioner validates that resource's config, so
`TestRelationalPathsResolveInSchema` (`services/all`) walks **every** registered resource's
relational paths against its schema — a bad path fails that test in CI, not at apply.

**2. `ConfigValidators` (reusable objects).** Plug prebuilt `resource.ConfigValidator`
values (e.g. from `terraform-plugin-framework-validators`, or a shared custom one). They
run after the `Relational`-derived validators. Use this when a rule is already packaged as
a composable, self-describing validator object.

**3. `ValidateConfig` (imperative).** A full framework `ValidateConfig` method for logic not
worth expressing as a validator object (e.g. Kusto's SKU-tier check). Runs after the base's
own work; values may be unknown at plan time, so guard for null/unknown before reading.

### Escape hatch — a framework interface `Base` doesn't implement

Hooks cover the common runtime seams (CRUD, the validation tiers above, `ModifyPlan`). Four
optional framework interfaces are **not** hooks — `ResourceWithIdentity`,
`ResourceWithUpgradeIdentity`, `ResourceWithUpgradeState`, `ResourceWithMoveState` (plus a
wholesale-different `ImportState`/`Create`/…). Go interface satisfaction is static, so a
single shared `Base` implementing e.g. `ResourceWithIdentity` would force identity onto
*every* native resource. Add them per-resource by wrapping `Base` and registering the
wrapper — `New` applies it, so the provider stays oblivious:

```go
type siteWithIdentity struct{ *resource.Base }
func (r *siteWithIdentity) IdentitySchema(ctx context.Context, _ fwresource.IdentitySchemaRequest, resp *fwresource.IdentitySchemaResponse) { /* … */ }
func init() {
    resource.RegisterOverride(WebSite.Name, func(b *resource.Base) fwresource.Resource {
        return &siteWithIdentity{Base: b}
    })
}
```

The embedded `*Base` methods stay promoted; the outer type's method set wins. Reserved for
the rare resource needing an interface `Base` lacks or a wholesale-different lifecycle method.

---

## 1. Add a new resource from the ground up

Worked example: `Microsoft.Storage/storageAccounts/blobServices` →
`azapi_storage_account_blob_service` (a singleton **child** of a storage account).

### Step 1 — Register the ARM type

```go
// internal/native/armtype/armtype.go
const StorageAccountBlobService = "Microsoft.Storage/storageAccounts/blobServices"
```

### Step 2 — Extract AzureRM knowledge with azwise

**Don't skip this** — without azwise the schema carries only bicep-derived flags and
loses curated ForceNew / validation / defaults / timeouts. Use the **azwise agent**
(`skill://azwise`), which drives `azwise_extract` over `terraform-provider-azurerm`:

1. `category=schema resource_name=azurerm_storage_account` → fields classified
   (computed/default/forceNew/sensitive/required/optional) + API version + blocks.
2. `category=automap` → ARM body paths per field.
3. `category=validation` → every `ValidateFunc` (enum/range/length/regex/UUID/custom).
   **Transfer all — never silently drop one.**
4. `category=timeouts` / `category=softdelete`.
5. `read`/`search` azurerm source only for paths automap couldn't verify.
6. Write **one** `internal/azure/azwise/<resource>.go` embedding `BaseKnowledge`:

```go
type StorageAccountBlobService struct{ BaseKnowledge }

var _ ResourceKnowledge = (*StorageAccountBlobService)(nil)

func NewStorageAccountBlobService() *StorageAccountBlobService {
    return &StorageAccountBlobService{BaseKnowledge: BaseKnowledge{
        ResourceType:   "Microsoft.Storage/storageAccounts/blobServices",
        ApiVersions:    []string{"2025-08-01"},
        TimeoutsConfig: &Timeouts{Create: 30 * time.Minute /* … */},
        StringRules:    []StringRule{ /* enum/regex/length, by ARM path */ },
        IntRules:       []IntRule{ /* numeric ranges */ },
        ArrayRules:     []ArrayRule{ /* MaxItems */ },
        DefaultValues:  []DefaultValue{ /* {PropertyPath, Value} */ },
        // ForceNew / ComputedFields / SensitiveFields / RequiredFields as needed
    }}
}
```

7. Register in `register.go`'s `RegisterAll()`: `Register(NewStorageAccountBlobService())`.

**Sub-service separation (critical).** Settings AzureRM bundles into a parent block —
`blob_properties`, `share_properties`, `queue_properties` in `azurerm_storage_account` —
are *separate* ARM resources (`…/blobServices/default`). Their knowledge goes in the
**sub-service's** file, never the parent's.

**Field-type discipline.** `StringRules` → string/enum fields, `IntRules` → numeric.
Cross-check the go-azure-sdk model struct; cite source file/line in the doc comment.

The generator overlays this via `ApplyAzwise` (Rule 9b).

### Step 3 — Add a generation target

```go
// internal/native/generator/cmd/generate_poc.go
var targets = []string{
    armtype.StorageAccount,
    armtype.StorageAccountBlobService, // ← add
}
```

Version + `types.json` path resolve from `index.json`; you name only the bare ARM type.

### Step 4 — Add a customizer (only if needed)

Skip when bicep + azwise suffice. A customizer is generation-time code that mutates the
intermediate `*typegraph.ResourceDefinition` for rules **neither bicep nor azwise can
express** — it is the last layer and wins over both. One file per resource; `register.go`
owns the single `init()` that wires it by ARM type:

```go
// internal/native/generator/customizers/storage_account_blob_service.go
func customizeStorageAccountBlobService(def *typegraph.ResourceDefinition) {
    // blobServices name is always "default" and isn't in the body graph → pin on envelope.
    def.SetNameValidators(
        typegraph.OneOfValidator(`blob service name must be "default"`, "default"),
    )
}
// register.go init(): Register(armtype.StorageAccountBlobService, customizeStorageAccountBlobService)
```

**The fluent vocabulary.** `*typegraph.ResourceDefinition` exposes intent-named,
path-variadic methods (in `internal/native/typegraph/customize.go`) that collapse the raw
`FindProperty(def, path)` + flip-a-flag dance into one call. Every method returns `def`, so
calls chain (`def.Required("sku", "sku.name").Default("properties.minimumTlsVersion", "TLS1_2")`).
Each targets **body properties by ARM dot path** (camelCase, the same paths `FindProperty`
understands) or the **operational envelope**. Prefer these over hand-mutating `*Property`
fields — the whole codebase uses them, and they carry the panic-on-typo guarantee.

Scenario → helper (body-property methods take one or more ARM paths, `paths ...string`):

| I need to… | Call | Under the hood |
|---|---|---|
| Attach a semantic validator (ID shape, regex, element enum) | `def.AddValidatorsFor(path, vs…)` | appends to `Property.Validators` |
| Promote a field AzureRM requires but bicep left optional | `def.Required(paths…)` | sets `FlagRequired` |
| Force a server-owned field to Computed-only | `def.Computed(paths…)` | `ForceComputed = true` |
| Mark an immutable field for replacement | `def.ForceNew(paths…)` | emits `RequiresReplace` |
| Mark a secret field sensitive | `def.Sensitive(paths…)` | emits `Sensitive: true` |
| Emit an unordered primitive array as a Set (kill reorder drift) | `def.AsSet(paths…)` | `UseSet` → `SetAttribute` |
| Default an omitted Optional+Computed list to `[]` (ARM echoes `[]`) | `def.WithEmptyListDefault(paths…)` | `DefaultEmptyList` |
| Let a null computed field plan as unknown so the server may fill it | `def.NonNullStateForUnknown(paths…)` | `UseNonNullStateForUnknown` |
| Attach an arbitrary plan modifier the built-ins can't express (drift normalization, conditional replace, bespoke) | `def.AddPlanModifiersFor(path, mods…)` | appends to `Property.PlanModifiers` |
| Same, on the resource **name** / **parent** ref | `def.AddNamePlanModifiers(mods…)` / `def.AddParentPlanModifiers(mods…)` | envelope attr `PlanModifiers` |
| Override a description-mined default | `def.Default(path, value)` | sets schema default |
| Constrain the resource **name** (not in the body graph) | `def.SetNameValidators(vs…)` | envelope name validators |
| Rename/redescribe the **parent** ref (e.g. `parent_id`→`scope_id`) | `def.SetParent(name, desc)` | envelope parent attr |
| Add a behavior-only top-level flag a hook reads (not sent to ARM) | `def.AddMetaAttr(attrs…)` | envelope Meta (Optional bool) |

Validator constructors (also `typegraph.*`, in `envelope.go`): `RegexValidator`,
`LengthValidator(min,max)` (negative bound = unbounded), `OneOfValidator`,
`OneOfCaseInsensitiveValidator`, `IntRangeValidator`, `ListSizeAtLeastValidator`,
`ListSizeAtMostValidator`, and `Validator(fn)` — which references a hand-written schema
validator **by its constructor symbol** (e.g. `typegraph.Validator(validators.UUID)`) so a
rename/delete is a compile error, not a silently wrong string. Reuse the generic
`validators.UUID` / `validators.AzureResourceID` from `internal/native/schema/validators`
before adding one; a rule tied to a single service (e.g.
`storagevalidators.StorageAccountIPRule`) lives in
`internal/native/services/<service>/validators`, imported under a `<svc>validators` alias. A
new validator is just a `validator.String` ctor in the package matching its scope — no
catalog, no registration; the emitter reflects the func to derive the call qualifier + import
alias.

**Array element shared with a sibling** (`ipRules` vs `ipv6Rules` dedupe to one bicep element
type) — call `typegraph.IsolateArrayElement(def, "<array-path>")` **first**, then address the
element property by its full path, otherwise the change leaks to the sibling array.

**A behavior-only Meta attribute pairs with its hook.** `def.AddMetaAttr` adds a top-level
flag that is *not* in the ARM body and never travels to ARM (e.g. `purge_on_destroy`); it is
emitted as an `Optional` bool the runtime hook reads from state. Ship the schema flag and the
consuming hook together. See GENERATOR.md Rule 9c "Meta attributes".

**Custom plan modifiers — by symbol, like validators.** The generator emits a *closed*
built-in set of plan modifiers (`UseStateForUnknown` / `UseNonNullStateForUnknown`,
`RequiresReplace`, the location and discriminated-variant modifiers). For anything else —
**drift normalization** (suppress a perpetual diff when Azure echoes a value in different
form, e.g. resource-ID casing), **conditional replace** on a custom predicate, or a bespoke
modifier — attach a hand-written one with `def.AddPlanModifiersFor(path, typegraph.PlanModifier(fn))`.
Like `Validator(fn)`, `PlanModifier(fn)` references the constructor **by symbol** (rename/delete
→ compile error); the emitter reflects it to the qualified call + import, and appends it **after**
the built-ins in the same typed `PlanModifiers` slice. The constructor must return the framework
plan-modifier type matching the attribute's kind (`planmodifier.String` for a string attr, `.Object`
for an object, …) — a mismatch is a compile error in the regenerated `_gen.go`. Reusable generic
modifiers (e.g. `nativeschema.UseStateForEquivalentResourceID`) live in `internal/native/schema`
(imported as `nativeschema`); a service-specific one goes in
`internal/native/services/<service>/planmodifiers`, imported under a `<svc>planmodifiers` alias.
This is a **schema** modifier baked into `_gen.go`; a *value-dependent* replace decision
(comparing old vs new, e.g. storage SKU zone-migration) is not a static modifier — it stays in a
`ModifyPlan` hook consulting `azwise.CheckForceNew`.

Everything above shapes the **schema**. If you find yourself wanting to change a value at
apply time, normalize a read, or gate on another attribute, that's a **hook**, not a
customizer — see [Runtime behavior hooks](#runtime-behavior-hooks).

`FindProperty` / `IsolateArrayElement` and every path-based method **panic** on a bad path —
fix the path, don't suppress.

### Step 5 — Regenerate

```bash
go run ./internal/native/generator/cmd/generate_poc.go
```

`<name>_gen.go` self-registers via `init()`. A **new service** → add one blank import to
`services/all/all.go`. No provider edits (`provider.go` iterates `services.Registry`).

### Step 6 — Validate

```bash
go run ./internal/native/cmd/azapin-validate/   # → "0 mismatches", no INVARIANT lines
```

### Step 7 — Test

Three surfaces (model on storage account / blob service):

| Test | File | Asserts |
|---|---|---|
| Runtime composition | `resource/base_test.go` (`TestBlobServiceSchemaComposition`) | schema composes (envelope + body + timeouts), parent attr resolves |
| azwise overlay | `generator/azwise_overlay_test.go` (`TestApplyAzwiseBlobService`) | the overlay baked into the live-parsed body |
| Acceptance | `services/<service>/<resource>_test.go` (a `Describe`) + `<resource>_config.go` (a `<Resource>Cfg` config builder) | live create/read/update against Azure |

**Native dependency rule (hard stop).** Acceptance configs for native resources must
depend only on other native generated resources. If a scenario needs a prerequisite ARM
resource and `services.Registry`/`internal/native/services/<service>` has no native
descriptor for it, stop the workflow: print a clear error naming the missing ARM type
and terminate generation/development for that resource until the dependency is added.
Never hide a missing native dependency by embedding generic `azapi_resource` HCL in a
native acceptance config; add the dependency as its own native target first, then use
its typed `<Dependency>Cfg` and `scope.ResourceFor` handle.

**Config builders** live in `<resource>_config.go` beside the schema. Split address
metadata from scenario HCL:

- `<Resource>Cfg` (e.g. `StorageAccountCfg`) embeds `services.ResourceConfigBase`, is
  constructed via `New<Resource>Cfg(parentCfgs..., optionalLabel)`, and carries only
  the Terraform type/label plus dependencies needed by every scenario. Pass this to
  `ws.ResourceFor(cfg)` / `scope.ResourceFor(cfg)` to get the acceptance handle.
- Each scenario is a named type implementing `nativeacc.Configure` (`Config() string`).
  Use a new defined type over the base config when the scenario needs no extra inputs
  (`type FooCfg_Basic FooCfg`); use a struct embedding the base only when the scenario
  has scenario-specific fields (e.g. `SKU string`, `Name string`, `Enabled bool`).
  Apply the scenario value, not a `Basic()` / `Update()` / `WithXXX()` method call.
- For ad-hoc raw HCL, wrap the string with `acc.StringConfigure(...)`; `Apply`,
  `Stage` and `ApplyExpectError` intentionally accept `Configure`, not `any`.

Example:

```go
cfg := storage.NewStorageAccountCfg(rgCfg)
sa := scope.ResourceFor(cfg)
sa.Apply(storage.StorageAccountCfg_Basic(cfg), acc.Exists())
sa.Apply(storage.StorageAccountCfg_Complete(cfg))
sa.Apply(storage.StorageAccountCfg_SKU{StorageAccountCfg: cfg, SKU: "Standard_ZRS"})
```

**Scenario convention**:
- `*_Basic` — minimal **create baseline**: required fields **plus a present (often
  empty `{}`) block for every all-optional nested object that carries computed/azwise
  defaults**. That block is load-bearing: the framework applies a nested default only
  when its parent object is non-null (`tftypes.Transform` does not descend into a null
  object), so omitting `properties = {}` silently drops every nested default (e.g.
  `minimum_tls_version=TLS1_2`) from the PUT body and the server keeps its own, often
  less safe, default.
- `*_Complete` — as many optional properties as one valid in-place configuration covers,
  exercising the whole resource surface.
- `*_Complete_update` — the same shape as `*_Complete` with **every in-place-updatable
  property set to a different valid value**, so applying `*_Complete` → `*_Complete_update`
  proves each one survives an in-place Update (Update → Read → empty plan) and that Azure
  accepts the new value. Hold a property equal to `*_Complete` (exclude it from the mutation
  set) only when it is genuinely **not** in-place-updatable here: a ForceNew / platform-identity
  field, one needing a dependency the suite doesn't provision, an enabling toggle held on so
  its dependent value is what changes, or a write-accepted-but-not-honored property (next note).
- Additional named scenarios encode their dependency/input explicitly in fields (e.g.
  `StorageAccountCfg_SKU{SKU: "Standard_ZRS"}` or
  `BlobServiceCfg_ChangeFeed{Enabled: false}`).
- Chain them in one `Ordered` container (`*_Basic` → intermediate scenarios →
  `*_Complete`) so a single resource is created once and mutated through each state,
  every `Apply` re-planning for drift. Keep ForceNew knobs (a replacing SKU, a
  `RequiresReplace` field) out of `*_Complete` and in their own scenario so the chain
  stays in-place. `tags` is modeled as an empty object, not a map — set what the schema
  accepts.

**Write-accepted-but-not-honored properties.** A property can be writable in the bicep flags
(`flags: 0`) yet the service normalizes or overrides it on GET, so a non-default value never
round-trips on a live resource — e.g. `Microsoft.Web/sites` `scmSiteAlsoStopped` (honored only
while the app is stopped) and `siteConfig.websiteTimeZone` (the `WEBSITE_TIME_ZONE` app setting
wins). AzureRM tends to omit such fields entirely. Never fabricate the value in an `AfterRead`
hook to hide the diff — that asserts state Azure contradicts (inconsistent-apply). Hold the
property at its round-tripping value in the config and record why in a comment.

**Template placeholders** — HCL Go-template fields rendered once per workspace
(`acceptance/render.go`, `tmplData`), fixed for its lifetime so the shared base and every
resource-under-test agree on names/placement:

| Placeholder | Type | Source | Use it for |
|---|---|---|---|
| `{{.RandomInteger}}` | `int` | `acceptance.RandTimeInt()` | unique numeric names, e.g. `acctest-rg-{{.RandomInteger}}` |
| `{{.RandomString}}` | `string` (5 lowercase alnum) | `acctest.RandStringFromCharSet` | tight length/charset names, e.g. `acctestsa{{.RandomString}}` |
| `{{.Location}}` | `string` | `ARM_TEST_LOCATION` (default `westeurope`), normalized | the primary region |
| `{{.LocationAlt}}` | `string` | `ARM_TEST_LOCATION_ALT` (default `eastus`), normalized | a second region for cross-region scenarios |
| `{{.SubscriptionID}}` | `string` | `ARM_SUBSCRIPTION_ID` | subscription-scoped IDs |

### Step 8 — Document

Refresh the DEVELOPER_SPEC.md "Current status" inventory (§7) if the resource set or a
component's status changed.

### New-resource checklist

- [ ] ARM type constant in `armtype.go`
- [ ] azwise `<resource>.go` written + `Register(New…())` in `register.go`
- [ ] Sub-service settings in the sub-service file, not the parent
- [ ] Target added to `generate_poc.go`
- [ ] Customizer added *only if* bicep + azwise can't express the rule
- [ ] New service → blank import in `services/all/all.go`
- [ ] Regenerate → `0 mismatches`
- [ ] Composition + azwise + acceptance tests
- [ ] Verification gate green

---

## 2. Update / regenerate an existing resource

Schema change **without** moving the API version: a corrected azwise rule, a new
validator, a customizer tweak, or a re-vendored `types.json` at the *same* version.

Golden rule: **never hand-edit `<name>_gen.go`** (it carries
`// Code generated by azapin; DO NOT EDIT.`). Edit the source layer, then regenerate.

| You want to change… | Edit this layer | Then |
|---|---|---|
| A ForceNew / default / validator / enum from AzureRM | azwise `<resource>.go` | regenerate |
| A name constraint, semantic validator, or rule bicep+azwise can't express | customizer `<resource>.go` | regenerate |
| A cross-resource validator | `internal/native/schema/validator_<rule>.go` | regenerate |
| Runtime behavior (mutate body before PUT, normalize on read) | `services/<service>/<name>_hooks.go` (`resource.RegisterHooks`) | no regenerate (runtime) |
| A cross-property constraint (`ExactlyOneOf`, `RequiredWith`, …) or complex validator | `services/<service>/<name>_hooks.go` (`Hooks.Relational` / `ConfigValidators` / `ValidateConfig`) | no regenerate (runtime) |

1. Edit the source layer.
2. Regenerate: `go run ./internal/native/generator/cmd/generate_poc.go`.
3. **Sanity-check the diff is intentional** — a *no-op* edit must leave `<name>_gen.go`
   byte-identical: `git diff --stat internal/native/services/`.
4. Run the verification gate. `azapin-validate` stays at `0 mismatches`; the azwise overlay
   test still asserts your rule baked in.

Runtime-only changes (hooks in `services/<service>/<name>_hooks.go`) need no regeneration —
just `go build` + tests.

---

## 3. Upgrade the API version of an existing resource

**Key fact:** generation always targets the **latest stable** (non-preview) version per
ARM type, resolved from `index.json` (GENERATOR.md Rule 16). No per-target pin — "upgrading"
means *making a newer stable version available*, then regenerating to follow it and
reconciling the overlay/customizer against the new body.

### Step 1 — Vendor newer bicep types

```bash
scripts/bicep-types-update.sh <path-to-bicep-types/generated-parent>
```

Wholesale-replaces `internal/azure/generated/` (all `types.json` + `index.json`);
`LatestStableVersion` then returns the new version (previews ignored). If nothing newer is
published upstream, there's nothing to upgrade to.

### Step 2 — Confirm latest-stable advanced

```bash
grep -o 'Microsoft.Storage/storageAccounts@[0-9-]*' internal/azure/generated/index.json | sort -u
```

The newest non-preview entry is what generation selects.

### Step 3 — Regenerate

```bash
go run ./internal/native/generator/cmd/generate_poc.go
```

`APIVersion` updates from the resolved tag; added/removed/renamed body properties flow through.

### Step 4 — Reconcile azwise (the silent layer)

`ApplyAzwise` resolves **exact version → catch-all (empty `ApiVersions`) → first registered
entry**, so the overlay keeps applying even when `ApiVersions` doesn't list the new version
(storage relies on this: `ApiVersions` is `["2025-08-01"]` while it generates at `2025-06-01`).
But a **renamed/moved** property means the path no longer resolves and the rule is **silently
dropped**. Reconcile:
- Run the overlay tests (parse live `types.json`, fail on a dropped rule):
  `go test ./internal/native/generator/ -run TestApplyAzwise`.
- Update the ARM dot path for anything that moved.
- Optionally add the new version to `ApiVersions` for an **exact** match (recommended once verified).

### Step 5 — Reconcile customizers (the loud layer)

A moved property makes `FindProperty` / `IsolateArrayElement` **panic** with the offending
path named, and regeneration fails — fix the path in the customizer.

### Step 6 — Validate

```bash
go run ./internal/native/cmd/azapin-validate/   # → "0 mismatches"
```

Catches body properties added/removed at the new version. Resolve every mismatch.

### Step 7 — Test

- Runtime composition (`resource/base_test.go`) — update expected attribute paths.
- Acceptance (`acceptance/<resource>_test.go`) — update body fields that changed shape; run
  live with `TF_ACC=1`.

### API-version-upgrade checklist

- [ ] Newer stable types vendored (`bicep-types-update.sh`); `index.json` advanced
- [ ] Regenerated; `<name>_gen.go` reflects the new version + body
- [ ] azwise overlay tests pass (no silently-dropped rule); paths/`ApiVersions` updated
- [ ] Customizer paths reconciled (no generation panic)
- [ ] `azapin-validate` → `0 mismatches`
- [ ] Composition + acceptance tests updated and green
- [ ] Verification gate green
