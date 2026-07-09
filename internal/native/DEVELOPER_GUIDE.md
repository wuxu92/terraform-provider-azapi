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
| ARM type constants | `internal/native/armtypes/armtypes.go` |
| azwise knowledge (curated AzureRM) | `internal/azure/azwise/<resource>.go` + `register.go` |
| Generation targets | `internal/native/generator/cmd/generate_poc.go` |
| Customizers (generation-time schema) | `internal/native/generator/customizers/<resource>.go` + `register.go` |
| Shared validators (cross-resource) | `internal/native/schema/validator_<rule>.go` |
| Per-service validators | `internal/native/services/<service>/validators/<rule>.go` |
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
| `BeforeDelete` | delete, before the DELETE call | preflight/guard using `ctx.State` |
| `AfterDelete` | delete, after the DELETE succeeds | destroy-time orchestration once the resource is gone (e.g. purge a Key Vault's soft-deleted shadow) using `ctx.State` |
| `ValidateConfig` | config validation (plan-time; values may be unknown) | cross-field rules one validator can't express |
| `ModifyPlan` | plan, **after** the base's default work | conditional `RequiresReplace`, plan-time derivation |
| `Singleton` (data) | create/delete of a fixed-named default child ARM never creates or deletes | skip the create existence check + reset to `DefaultBody` via PUT on destroy (e.g. `blobServices/default`); the reset path runs neither `BeforeDelete` nor `AfterDelete` |

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

---

## 1. Add a new resource from the ground up

Worked example: `Microsoft.Storage/storageAccounts/blobServices` →
`azapi_storage_account_blob_service` (a singleton **child** of a storage account).

### Step 1 — Register the ARM type

```go
// internal/native/armtypes/armtypes.go
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
    armtypes.StorageAccount,
    armtypes.StorageAccountBlobService, // ← add
}
```

Version + `types.json` path resolve from `index.json`; you name only the bare ARM type.

### Step 4 — Add a customizer (only if needed)

Skip when bicep + azwise suffice. Add a customizer for rules **neither can express**.
One file per resource; `register.go` owns the single `init()`:

```go
// internal/native/generator/customizers/storage_account_blob_service.go
func customizeStorageAccountBlobService(def *generator.ResourceDefinition) {
    // blobServices name is always "default" and isn't in the body graph → pin on envelope.
    def.Envelope.Name.Validators = []generator.DescriptionValidator{
        generator.OneOfValidator(`blob service name must be "default"`, "default"),
    }
}
// register.go init(): Register(armtypes.StorageAccountBlobService, customizeStorageAccountBlobService)
```

Customizers mutate `*Property` by ARM dot path and run **last** (win over azwise +
description-mining). Common moves:
- **Body property** — `FindProperty(def, "properties.minimumTlsVersion").DefaultValue = "TLS1_2"`
  (also sets `ForceNew` / `Sensitive` / `Validators`).
- **Envelope name** — `def.Envelope.Name.Validators = …` (name isn't in the body).
- **Semantic validator** — attach a `validator.String`, home by reusability:
  - Generic/cross-resource → `SharedValidator("UUID()")`, ctor in
    `schema/validator_<rule>.go`. Reuse `UUID()`, `AzureResourceID()`, … before adding one.
  - Resource-specific → `CustomValidator("StorageAccountIPRule()")`, ctor in
    `services/<service>/validators/<rule>.go`.
- **Array element shared with a sibling** (`ipRules` vs `ipv6Rules`) — call
  `IsolateArrayElement(def, "<path>")` **first** so the validator doesn't leak to the sibling.
- **Behavior-only Meta attribute** — `def.Envelope.Meta = append(def.Envelope.Meta, typegraph.MetaAttr{Name: "purge_on_destroy", Description: …})`
  for a top-level flag that is *not* in the ARM body and drives provider-side
  behavior only (emitted as an `Optional` bool; read by a runtime hook). Pair it with
  the hook that consumes it — schema flag and behavior ship together. See GENERATOR.md
  Rule 9c "Meta attributes".

`FindProperty` / `IsolateArrayElement` **panic** on a bad path — fix the path, don't suppress.

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

- [ ] ARM type constant in `armtypes.go`
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
