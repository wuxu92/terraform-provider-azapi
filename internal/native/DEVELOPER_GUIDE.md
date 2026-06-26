# Azapin Developer Guide

Task-oriented playbook for azapin static resources. Three workflows:

1. [Add a new resource from the ground up](#1-add-a-new-resource-from-the-ground-up) — incl. azwise knowledge extraction.
2. [Update / regenerate an existing resource](#2-update--regenerate-an-existing-resource) — bicep/azwise/customizer changed, same API version.
3. [Upgrade the API version of an existing resource](#3-upgrade-the-api-version-of-an-existing-resource).

Reference docs cover the *why*; this covers the *how*:
- **DESIGN.md** — architecture, naming, scope.
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

The result is baked into `internal/native/generated/<service>/<name>_gen.go`. The
**runtime never mutates the schema** — it serves the generated `Schema()` plus a
`timeouts` block. Per-resource runtime *behavior* (not schema) is a separate hook
layer (`resource/overlay_<name>.go`); see [Runtime behavior hooks](#runtime-behavior-hooks).

Two failure modes to internalize:
- **Customizer paths panic.** `FindProperty` / `IsolateArrayElement` crash generation
  on an unresolvable ARM path — a typo or renamed property can't ship silently.
- **azwise paths fail silently.** `ApplyAzwise` skips an unresolved path; a *lost* rule
  is caught only by the overlay tests (`TestApplyAzwise*`), which parse live
  `types.json` and assert the rule baked in.

## Where things live

| What | Path |
|---|---|
| ARM type constants | `internal/native/armtypes/armtypes.go` |
| azwise knowledge (curated AzureRM) | `internal/azure/azwise/<resource>.go` + `register.go` |
| Generation targets | `internal/native/generator/cmd/generate_poc.go` |
| Customizers (generation-time schema) | `internal/native/generator/customizers/<resource>.go` + `register.go` |
| Shared validators (cross-resource) | `internal/native/schema/validator_<rule>.go` |
| Per-service validators | `internal/native/generated/<service>/validators/<rule>.go` |
| Generated output | `internal/native/generated/<service>/<name>_gen.go` |
| Service aggregator (blank imports) | `internal/native/generated/all/all.go` |
| Runtime base / mapper / hooks | `internal/native/resource/` |
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
TF_ACC=1 go test ./internal/native/generated/<service>/ -run <TestXxxAcceptance>
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

Register per Terraform name from an overlay `init()`. Overlays live in package
`resource`, so they compile in automatically — **no blank import**:

```go
// internal/native/resource/overlay_<name>.go
package resource

func init() {
    RegisterHooks("azapi_<name>", &Hooks{BeforeCreate: ..., ModifyPlan: ...})
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
| `AfterRead` | read, after GET, before mapping into state | massage `ctx.Response` before flatten |
| `BeforeDelete` | delete, before the DELETE call | preflight/guard using `ctx.State` |
| `ValidateConfig` | config validation (plan-time; values may be unknown) | cross-field rules one validator can't express |
| `ModifyPlan` | plan, **after** the base's default work | conditional `RequiresReplace`, plan-time derivation |

`Before/AfterCreate` vs `Before/AfterUpdate` dispatch by whether the op is a create
(same body-composition path). `ValidateConfig`/`ModifyPlan` use framework signatures and
run *after* the base (base applies schema `RequiresReplace` first, then your `ModifyPlan`).

> `Hooks.BeforeRead` exists in the struct but `Read` invokes only `AfterRead` —
> `BeforeRead` is a **no-op today**. Wire it in `base.go`'s `Read` before relying on it.

### What a hook sees — `CrudCtx`

| Field | What | Notes |
|---|---|---|
| `Ctx` | operation `context.Context` | already timeout-scoped |
| `Client` | `*clients.Client` | make extra ARM calls if needed |
| `ID` | `parse.ResourceId` | the resource's Azure ID |
| `Plan` | typed plan object | set on create/update, null otherwise |
| `State` | typed prior-state object | set on read/update/delete, null otherwise |
| `Body` | `map[string]interface{}` — ARM body being composed | **mutate in `Before*`** |
| `Response` | `map[string]interface{}` — ARM GET response | **read in `After*`** |
| `Diags` | `*diag.Diagnostics` | append an error + `return` to abort |

`Before*` mutate `Body`; `After*` read `Response`. The base checks `HasError()` after
each hook — append a diagnostic and `return` to stop the operation.

### Example — conditional ForceNew (shipped)

A static plan modifier can't express "replace only when migrating between zonal and
non-zonal SKUs", so it lives in a `ModifyPlan` hook consulting `azwise.CheckForceNew`:

```go
// internal/native/resource/overlay_storage_account.go
func storageAccountModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
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
    // ARMType/APIVersion from the shipped descriptor — stays correct across version upgrades.
    d := generated.Registry["azapi_storage_account"]
    oldBody := map[string]interface{}{"sku": map[string]interface{}{"name": oldName.ValueString()}}
    newBody := map[string]interface{}{"sku": map[string]interface{}{"name": newName.ValueString()}}
    if azwise.CheckForceNew(d.ARMType, d.APIVersion, oldBody, newBody) {
        resp.RequiresReplace = append(resp.RequiresReplace, path.Root("sku"))
    }
}
```

### Example — body / response massaging (template)

```go
RegisterHooks("azapi_<name>", &Hooks{
    BeforeCreate: func(c *CrudCtx) { c.Body["someField"] = derive(c) },        // injected, then PUT
    AfterRead:    func(c *CrudCtx) { /* massage c.Response before flatten, or c.Diags.AddError(…) */ },
})
```

### Hook rules

- Schema concerns (validators, defaults, ForceNew on a fixed attribute) → generated
  schema via customizers/azwise, **never** a hook.
- Keyed by the **Terraform** name; same-package `init()`, no import wiring — just drop
  `overlay_<name>.go` into `resource/`.
- Read ARM type/version from `generated.Registry[name]`, never hardcode.
- Append to `Diags`/`resp.Diagnostics` and `return` to abort; the base checks after each hook.

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
    `generated/<service>/validators/<rule>.go`.
- **Array element shared with a sibling** (`ipRules` vs `ipv6Rules`) — call
  `IsolateArrayElement(def, "<path>")` **first** so the validator doesn't leak to the sibling.

`FindProperty` / `IsolateArrayElement` **panic** on a bad path — fix the path, don't suppress.

### Step 5 — Regenerate

```bash
go run ./internal/native/generator/cmd/generate_poc.go
```

`<name>_gen.go` self-registers via `init()`. A **new service** → add one blank import to
`generated/all/all.go`. No provider edits (`provider.go` iterates `generated.Registry`).

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
| Acceptance | `generated/<service>/<resource>_test.go` (a `Describe`) + `<resource>_config.go` (a `<Resource>Cfg` config builder) | live create/read/update against Azure |

**Config builders** live in `<resource>_config.go` beside the schema: a `<Resource>Cfg`
struct (e.g. `StorageAccountCfg`) embedding `generated.ResourceConfigBase`, constructed via
`New<Resource>Cfg(label, …parentCfgs)`, exposing `Basic()` / `Update()` / `Complete()` /
`Named(…)` that return HCL; the Terraform type is set once from the resource's generated
`Descriptor` (e.g. `StorageAccount.Name`, the `var` the `_gen.go` exposes and registers).
Tests pass a config to `ws.ResourceFor(cfg)` / `scope.ResourceFor(cfg)` for the
handle they apply. **One Ginkgo `RunSpecs` per service** (`generated/<service>/suite_test.go`,
e.g. `TestStorageAcceptance`); add a `Describe(...)` in a new `<resource>_test.go`, never a
second `RunSpecs` (Ginkgo panics on more than one per binary).

**Config method convention** (mirrors azurerm):
- `Basic()` — minimal **create baseline**: required fields **plus a present (often empty
  `{}`) block for every all-optional nested object that carries computed/azwise defaults**.
  That block is load-bearing: the framework applies a nested default only when its parent
  object is non-null (`tftypes.Transform` does not descend into a null object), so omitting
  `properties = {}` silently drops every nested default (e.g. `minimum_tls_version=TLS1_2`)
  from the PUT body and the server keeps its own, often less safe, default.
- `Complete()` — as many optional properties as one valid in-place configuration covers,
  exercising the whole resource surface.
- `Update()` — an intermediate state that proves the update path.
- Chain them in one `Ordered` container (`Basic` → `Update` → `Complete`) so a single
  resource is created once and mutated through each state, every `Apply` re-planning for
  drift. Keep ForceNew knobs (a replacing SKU, a `RequiresReplace` field) out of
  `Update`/`Complete` and in their own scenario so the chain stays in-place. `tags` is
  modeled as an empty object, not a map — set what the schema accepts.

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

Refresh the DESIGN.md "Current State" inventory if the resource set or a component's
status changed.

### New-resource checklist

- [ ] ARM type constant in `armtypes.go`
- [ ] azwise `<resource>.go` written + `Register(New…())` in `register.go`
- [ ] Sub-service settings in the sub-service file, not the parent
- [ ] Target added to `generate_poc.go`
- [ ] Customizer added *only if* bicep + azwise can't express the rule
- [ ] New service → blank import in `generated/all/all.go`
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
| Runtime behavior (mutate body before PUT, normalize on read) | `resource/overlay_<name>.go` (`RegisterHooks`) | no regenerate (runtime) |

1. Edit the source layer.
2. Regenerate: `go run ./internal/native/generator/cmd/generate_poc.go`.
3. **Sanity-check the diff is intentional** — a *no-op* edit must leave `<name>_gen.go`
   byte-identical: `git diff --stat internal/native/generated/`.
4. Run the verification gate. `azapin-validate` stays at `0 mismatches`; the azwise overlay
   test still asserts your rule baked in.

Runtime-only changes (hooks in `overlay_<name>.go`) need no regeneration — just `go build`
+ tests.

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
