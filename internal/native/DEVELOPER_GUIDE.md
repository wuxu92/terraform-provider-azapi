# Azapin Developer Guide

This is the **task-oriented playbook** for working on azapin static resources. It
covers three workflows end to end:

1. [Add a new native resource from the ground up](#1-add-a-new-resource-from-the-ground-up) — including extracting AzureRM knowledge with azwise.
2. [Update / regenerate an existing resource](#2-update--regenerate-an-existing-resource) — when bicep types, azwise knowledge, or a customizer changes.
3. [Upgrade the API version of an existing resource](#3-upgrade-the-api-version-of-an-existing-resource).

It complements — not replaces — the reference docs:

- **DESIGN.md** — architecture, naming, project scope.
- **GENERATOR.md** — the 17 generator rules (flags, defaults, validators, envelope, customizers).
- **RESOURCE.md** — the runtime `Base`, the mapper, and hooks.

Read those for *why*; read this for *how*.

## Mental model

A generated resource is assembled at **generation time** from four layered inputs,
each able to override the one before it:

```
bicep flags (types.json)         base: Required/ReadOnly/WriteOnly → schema flags
  └─ description mining          defaults & validators scraped from ARM descriptions
       └─ azwise overlay         curated AzureRM knowledge (authoritative; Rule 9b)
            └─ customizers        hand-written Go, runs last, final say (Rule 9d)
```

The result is baked into `internal/native/generated/<service>/<name>_gen.go`. The
**runtime never mutates the schema** — it serves the generated `Schema()` and adds
only a `timeouts` block. Per-resource runtime *behavior* (not schema) is a separate
hook layer (`resource/overlay_<name>.go`).

Two failure modes worth internalizing before you start:

- **Customizer paths fail loudly.** `generator.FindProperty` / `IsolateArrayElement`
  **panic** on an ARM path that doesn't resolve — a typo or a renamed property
  crashes generation immediately. Good: you can't ship a stale customizer.
- **azwise paths fail silently.** `ApplyAzwise` **skips** a path that doesn't
  resolve in the body (so ListKeys-only fields, etc. are harmless). The net for a
  *lost* rule is the azwise overlay tests (`TestApplyAzwise*`), which parse the
  live `types.json` and assert the overlay actually baked in.

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

Every workflow ends with the same gate. Run it from the repo root:

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
TF_ACC=1 go test ./internal/native/acceptance/ -run <ResourceName>
```

`azapin-validate` is the load-bearing check: it confirms the emitted schema covers
every bicep body property and vice versa. `INVARIANT …` lines mean a `Default`
violates the framework's `Optional+Computed`/validator rules (GENERATOR.md
"Schema Flag Invariants") — fix before shipping.

---

## 1. Add a new resource from the ground up

Worked example: `Microsoft.Storage/storageAccounts/blobServices` →
`azapi_storage_account_blob_service` (a singleton **child** of a storage account).

### Step 1 — Register the ARM type

Add the bare ARM type (no API version) to the catalog so it is spelled once and
referenced everywhere by constant:

```go
// internal/native/armtypes/armtypes.go
const StorageAccountBlobService = "Microsoft.Storage/storageAccounts/blobServices"
```

### Step 2 — Extract AzureRM knowledge with azwise

**Do not skip this.** Without azwise the schema carries only bicep-derived flags and
loses curated ForceNew / validation / defaults / timeouts.

Use the **azwise agent** (`.omp/agents/azwise.md`, skill `skill://azwise`), which
drives the `azwise_extract` tool over the `terraform-provider-azurerm` source. Its
workflow (see the agent file for the full contract):

1. `azwise_extract category=schema resource_name=azurerm_storage_account` → fields
   classified computed/default/forceNew/sensitive/required/optional + API version + blocks.
2. `azwise_extract category=automap …` → ARM body paths for each field.
3. `azwise_extract category=validation …` → every `ValidateFunc` (enums, ranges,
   lengths, regexes, UUID, custom). **Transfer all of them — never silently drop one.**
4. `category=timeouts` / `category=softdelete`.
5. `read`/`search` azurerm source only for paths automap couldn't verify.
6. Write **one** `internal/azure/azwise/<resource>.go` embedding `BaseKnowledge`:

```go
// internal/azure/azwise/storage_account_blob_service.go
type StorageAccountBlobService struct{ BaseKnowledge }

var _ ResourceKnowledge = (*StorageAccountBlobService)(nil)

func NewStorageAccountBlobService() *StorageAccountBlobService {
    return &StorageAccountBlobService{
        BaseKnowledge: BaseKnowledge{
            ResourceType:   "Microsoft.Storage/storageAccounts/blobServices",
            ApiVersions:    []string{"2025-08-01"},
            TimeoutsConfig: &Timeouts{Create: 30 * time.Minute, /* … */},
            StringRules:    []StringRule{ /* enum/regex/length, by ARM path */ },
            IntRules:       []IntRule{ /* numeric ranges */ },
            ArrayRules:     []ArrayRule{ /* MaxItems */ },
            DefaultValues:  []DefaultValue{ /* {PropertyPath, Value} */ },
            // ForceNew / ComputedFields / SensitiveFields / RequiredFields as needed
        },
    }
}
```

7. Register it:

```go
// internal/azure/azwise/register.go — RegisterAll()
Register(NewStorageAccountBlobService())
```

**Sub-service API separation (critical).** Settings AzureRM bundles into a parent
block — `blob_properties`, `share_properties`, `queue_properties` inside
`azurerm_storage_account` — are *separate* ARM resources
(`…/blobServices/default`, …). Their knowledge goes in the **sub-service's** file,
never the parent's.

**Field-type discipline.** `StringRules` target string/enum fields, `IntRules`
target numeric fields. Cross-check the go-azure-sdk model struct; cite source
file/line in the struct doc comment.

The generator overlays this automatically via `ApplyAzwise` (Rule 9b) — declarative
rules become baked validators/defaults/plan-modifiers.

### Step 3 — Add a generation target

```go
// internal/native/generator/cmd/generate_poc.go
var targets = []string{
    armtypes.StorageAccount,
    armtypes.StorageAccountBlobService, // ← add
}
```

The version and `types.json` path are resolved from `index.json` automatically — you
name only the bare ARM type.

### Step 4 — Add a customizer (only if needed)

Skip this when bicep + azwise already suffice. Add a customizer for rules **neither
bicep nor azwise can express**. Each resource gets its own file; `register.go` owns
the single `init()`:

```go
// internal/native/generator/customizers/storage_account_blob_service.go
func customizeStorageAccountBlobService(def *generator.ResourceDefinition) {
    // blobServices is a singleton: ARM name is always "default", and name is not
    // in the body graph — so pin it on the envelope name attribute.
    def.Envelope.Name.Validators = []generator.DescriptionValidator{
        generator.OneOfValidator(`blob service name must be "default" …`, "default"),
    }
}
```

```go
// internal/native/generator/customizers/register.go — init()
func init() {
    Register(armtypes.StorageAccountBlobService, customizeStorageAccountBlobService)
}
```

Customizers mutate `*Property` fields by ARM dot path and run **last**, so they win
over azwise and description-mining. Common moves:

- **Body property** — `generator.FindProperty(def, "properties.minimumTlsVersion").DefaultValue = "TLS1_2"`
  (sets `DefaultValue` / `ForceNew` / `Sensitive` / `Validators`).
- **Envelope name** — `def.Envelope.Name.Validators = …` (name isn't in the body).
- **Semantic validator the declarative rules can't express** — attach a
  `validator.String`, picking its home by reusability:
  - **Generic / cross-resource** → `generator.SharedValidator("UUID()")`, constructor
    in `internal/native/schema/validator_<rule>.go` (emitted `nativeschema.UUID()`).
    Reuse `UUID()`, `AzureResourceID()`, … before writing a new one.
  - **Resource-specific** → `generator.CustomValidator("StorageAccountIPRule()")`,
    constructor in `internal/native/generated/<service>/validators/<rule>.go`
    (emitted `validators.StorageAccountIPRule()`).
- **Array element shared with a sibling** (e.g. `ipRules` vs `ipv6Rules`) — call
  `generator.IsolateArrayElement(def, "<array path>")` **first** so the validator
  doesn't leak to the sibling.

`FindProperty` / `IsolateArrayElement` **panic** on a bad path — that is the
intended guard; fix the path, don't suppress it.

### Step 5 — Regenerate

```bash
go run ./internal/native/generator/cmd/generate_poc.go
```

The new `<name>_gen.go` self-registers via `init()`. If this is a **new service**,
add one blank import to `internal/native/generated/all/all.go` so the registry is
populated. No provider edits — `provider.go` iterates `generated.Registry`.

### Step 6 — Validate

```bash
go run ./internal/native/cmd/azapin-validate/   # → "0 mismatches", no INVARIANT lines
```

### Step 7 — Test

Add three test surfaces (model them on the storage account / blob service tests):

| Test | File | Asserts |
|---|---|---|
| Runtime composition | `resource/base_test.go` (`TestBlobServiceSchemaComposition`) | schema composes (envelope + body + timeouts), parent attr resolves |
| azwise overlay | `generator/azwise_overlay_test.go` (`TestApplyAzwiseBlobService`) | the overlay baked into the live-parsed body |
| Acceptance | `acceptance/<resource>_test.go` (a `Describe` block) | live create/read/update against Azure |

> **Acceptance suites use one Ginkgo `RunSpecs`** (`acceptance/suite_test.go`). Add a
> `Describe(...)` block in a new `<resource>_test.go` — do **not** add a second
> `RunSpecs` / `TestXxx` entry point (Ginkgo panics on more than one per binary).

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

Use this when the schema needs to change **without** moving the API version: a
corrected azwise rule, a new validator, a customizer tweak, or a re-vendored
`types.json` at the *same* version.

The golden rule: **never hand-edit `<name>_gen.go`** (it carries
`// Code generated by azapin; DO NOT EDIT.`). Edit the *source layer*, then
regenerate.

| You want to change… | Edit this layer | Then |
|---|---|---|
| A ForceNew / default / validator / enum from AzureRM | azwise `<resource>.go` | regenerate |
| A name constraint, semantic validator, or a rule bicep+azwise can't express | customizer `<resource>.go` | regenerate |
| A cross-resource validator | `internal/native/schema/validator_<rule>.go` | regenerate |
| Runtime behavior (mutate body before PUT, normalize on read) | `resource/overlay_<name>.go` (`RegisterHooks`) | no regenerate (runtime) |

Workflow:

1. Edit the appropriate source layer above.
2. Regenerate: `go run ./internal/native/generator/cmd/generate_poc.go`.
3. **Sanity-check the diff is intentional.** A source change should move
   `<name>_gen.go` in exactly the way you expect; a *no-op* edit must leave it
   **byte-identical**:
   ```bash
   git diff --stat internal/native/generated/   # only the files you meant to change
   ```
4. Run the verification gate. `azapin-validate` must stay at `0 mismatches`; the
   azwise overlay test must still assert your rule baked in.

Note on runtime-only changes: hooks in `resource/overlay_<name>.go` (registered via
`RegisterHooks`) customize CRUD *behavior*, not schema, so they need no regeneration
— just `go build` + tests.

---

## 3. Upgrade the API version of an existing resource

**Key fact:** generation always targets the **latest stable** (non-preview) version
for each ARM type, resolved from `index.json` (`generator.Index`, GENERATOR.md
Rule 16). There is no per-target version pin. So "upgrading" means *making a newer
stable version available*, then regenerating to follow it — and reconciling the
overlay/customizer against the new body.

### Step 1 — Vendor newer bicep types

The embedded manifest is the single source of version truth. Refresh it from an
updated `bicep-types-az` checkout:

```bash
scripts/bicep-types-update.sh <path-to-bicep-types/generated-parent>
```

This wholesale-replaces `internal/azure/generated/` (all `types.json` + `index.json`).
After it runs, `index.json` lists the new version for the type, and
`LatestStableVersion` will return it (preview versions are ignored).

> If a newer version isn't published upstream yet, there is nothing to upgrade to —
> the generator already pins the newest stable that exists.

### Step 2 — Confirm latest-stable advanced

```bash
grep -o 'Microsoft.Storage/storageAccounts@[0-9-]*' internal/azure/generated/index.json | sort -u
```

The newest non-preview entry is what generation will select.

### Step 3 — Regenerate

```bash
go run ./internal/native/generator/cmd/generate_poc.go
```

The descriptor's `APIVersion` updates automatically (it comes from the resolved tag).
Expect `<name>_gen.go` to change: added/removed/renamed body properties flow through.

### Step 4 — Reconcile azwise (the silent layer)

`ApplyAzwise` calls `azwise.Get(armType, newVersion)`, which resolves
**exact version → catch-all (empty `ApiVersions`) → first registered entry**. So the
overlay keeps applying even when `ApiVersions` doesn't list the new version (storage
already relies on this — its `ApiVersions` is `["2025-08-01"]` while it generates at
`2025-06-01`).

But a property **renamed or moved** in the new body means the azwise path no longer
resolves and the rule is **silently dropped**. Reconcile:

- Run the azwise overlay tests — they parse the *live* `types.json` and assert each
  rule baked in, so a dropped rule fails the test:
  ```bash
  go test ./internal/native/generator/ -run TestApplyAzwise
  ```
- For any path that moved, update the ARM dot path in the azwise `<resource>.go`.
- Optionally add the new version to `ApiVersions` so `Get` returns an **exact** match
  and coverage is explicit (recommended once you've verified the overlay against it).

### Step 5 — Reconcile customizers (the loud layer)

If a customizer references a property that moved, `FindProperty` / `IsolateArrayElement`
**panic** and regeneration fails with the offending path named — fix the path in the
customizer. (This is why customizer breakage can't ship silently.)

### Step 6 — Validate

```bash
go run ./internal/native/cmd/azapin-validate/   # → "0 mismatches"
```

This catches body properties added or removed at the new version (schema ↔ bicep
parity). Resolve every mismatch before continuing.

### Step 7 — Test

Renamed/added properties commonly break assertions:

- Runtime composition (`resource/base_test.go`) — update expected attribute paths.
- Acceptance (`acceptance/<resource>_test.go`) — update body fields that changed
  shape; run live with `TF_ACC=1`.

### API-version-upgrade checklist

- [ ] Newer stable types vendored (`bicep-types-update.sh`); `index.json` advanced
- [ ] Regenerated; `<name>_gen.go` reflects the new version + body
- [ ] azwise overlay tests pass (no silently-dropped rule); paths/`ApiVersions` updated
- [ ] Customizer paths reconciled (no generation panic)
- [ ] `azapin-validate` → `0 mismatches`
- [ ] Composition + acceptance tests updated and green
- [ ] Verification gate green
