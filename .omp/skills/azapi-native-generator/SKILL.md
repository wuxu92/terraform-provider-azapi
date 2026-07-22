---
name: azapi-native-generator
description: azapin native static resource work — take a prose description of the wanted Azure resource, resolve it to an ARM type, and add, regenerate, or API-version-upgrade a typed azapi_* resource through the azwise overlay, customizers, hooks, and the verification gate. Use when a task touches an internal/native resource, describes a resource to generate natively, or when azwise knowledge needs a generated resource around it.
---

# azapin native static resource generation

An azapin resource is **generated, never hand-written**. It is built as a four-layer **overlay**, each layer overriding the one before:

```text
bicep types.json  ->  description mining  ->  azwise overlay  ->  customizer  ->  <name>_gen.go
```

`azwise` is the **silent layer** (an unresolved ARM path is skipped without error); the customizer is the **loud layer** (a bad path panics generation). Runtime hooks sit outside the overlay — they shape behavior, never schema shape.

The repo docs are the single source of truth; this skill orchestrates them. Read the one your branch names:

| Doc | Owns |
| --- | --- |
| `internal/native/DEVELOPER_GUIDE.md` | the three workflows, file map, verification gate, test conventions — **follow it step by step** |
| `internal/native/GENERATOR.md` | overlay rules, schema flag invariants |
| `internal/native/RESOURCE.md` | runtime base, mapper, hooks, data source |
| `internal/native/DEVELOPER_SPEC.md` | architecture, naming, scope, rollout status |
| `internal/native/SCHEMA_VERIFICATION.md` | live-API verification (typed resources are not discovery probes) |
| `skill://azwise` | writing the azwise overlay layer |

## Intake — resolve the description to a target

The user describes the resource in prose ("a native blob service", "typed Cosmos DB account", "expose Microsoft.Web/sites"). Before any branch runs, turn that into a concrete **target** and confirm it exists — a branch that starts from a wrong or unpublished type wastes the whole run.

1. **ARM type.** Map the description to a fully-qualified `<Namespace>/<type>`. If the prose is ambiguous (a product name spanning several ARM types, a parent block that is really a child resource — `blob_properties` → `Microsoft.Storage/storageAccounts/blobServices`), resolve it from the description and the sub-service rules in `DEVELOPER_GUIDE.md`; ask the user only when two genuinely distinct ARM types remain.
2. **Confirm it is published.** The type must resolve to a latest **stable** (non-preview) version in the embedded bicep index — this is also the version the generator will target:
   ```bash
   grep -oE '<Namespace>/<type>@[0-9][0-9-]*' internal/azure/generated/index.json | sort -u
   ```
   The newest non-preview tag is what generation selects. No stable entry → **stop and report** (preview-only types stay on `azapi_resource`).
3. **Derive name + home** mechanically, don't invent: Terraform name via `naming.ResourceName`, service package via `naming.ServiceName`, latest version via `azure.GetLatestStableApiVersion`. State the resolved `armType`, `azapi_*` name, service, and version back to the user before proceeding.

Completion criterion: a single ARM type, confirmed in the index, with its name/service/version resolved. Pick the branch only once you have it.

## Branch router

Pick one; each names the `DEVELOPER_GUIDE.md` section to execute in full.

- **Add a resource** → §1. The agent-critical spine:
  1. ARM type constant in `armtype/armtype.go`.
  2. **azwise overlay** — invoke `skill://azwise` (`azwise_extract` schema/automap/validation/relational/timeouts, then `azwise_validate`). Never skip; without it the schema loses curated lifecycle, defaults, validation, sensitive/computed, timeouts.
  3. Target in `generator/cmd/main.go`.
  4. Customizer **only** for a rule the overlay cannot express.
  5. Hooks **only** for runtime behavior.
  6. Regenerate; a new service package adds one blank import to `services/all/all.go`.
  7. Tests: composition + overlay offline; author acceptance configs but **do not run them live**.
  8. Verification gate.
- **Regenerate an existing resource** → §2. Edit the source layer that owns the change (azwise / customizer / validator / hook), then regenerate. A no-op source edit must leave `*_gen.go` byte-identical.
- **API-version upgrade** → §3. Regenerate against newer stable types, then reconcile both layers: overlay tests catch **silent** azwise path drops; a customizer panic is the **loud** signal to fix a moved path.

## Guardrails

Load-bearing where an agent tends to drift:

- Never hand-edit a `*_gen.go` file (`Code generated … DO NOT EDIT`). Edit a source layer, regenerate.
- Never run live acceptance (`TF_ACC=1`) as an AI agent — author and update the configs/tests, stop there.
- Never fix a `FindProperty` / `IsolateArrayElement` panic by suppressing it — fix the ARM path.
- Never put a schema concern (validator, default, ForceNew, sensitive, computed) in a hook — it belongs in the overlay.
- Never reach a missing native acceptance dependency through generic `azapi_resource` HCL — add it as its own native target first, or stop and report.
- Keep AzureRM sub-service settings (e.g. `blob_properties`) on the child ARM resource, never the parent.

## Stop and report when

- No stable non-preview API version exists for the ARM type.
- A required native acceptance dependency is absent and adding it is out of scope.
- The body is polymorphic/discriminated beyond what native typing models safely.
- No ARM path verifies from automap, source, SDK models, or bicep types.

## Verification

Run `DEVELOPER_GUIDE.md`'s verification gate. Focused offline subset an agent runs directly:

```bash
go run ./internal/native/generator/cmd/
go run ./internal/native/cmd/azapin-validate/            # 0 mismatches, no INVARIANT lines
go test ./internal/native/generator/ -run TestApplyAzwise -count=1
go test ./internal/native/resource/ -run Test.*SchemaComposition -count=1
go test ./internal/native/services/<service>/ -count=1   # azwise knowledge lives in github.com/wuxu92/azwise (separate module)
```

## Report

ARM type, Terraform name, service, API version · files changed per overlay layer · verification commands run and their result · any accepted limitation (skipped validator + reason, write-not-honored property, unrun live acceptance).
