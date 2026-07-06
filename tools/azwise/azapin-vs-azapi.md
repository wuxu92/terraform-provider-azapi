# azapin (static resources) vs `azapi_resource` (dynamic)

**Audience:** teammates and leadership.
**One-liner:** azapin turns a single freeform `azapi_resource` (raw ARM JSON) into a
typed, per-resource Terraform resource generated from Azure's own schema — with
native-schema autocomplete and per-attribute validation, correct diffs, and
AzureRM-grade operational knowledge — while `azapi_resource` stays the day-zero
escape hatch. (`azapi_resource` already validates its `body` against the same ARM
types at plan time; azapin makes that a first-class typed schema.)

Both ship in the **same provider**, share the same authentication and backend, and
are meant to **coexist**.

---

## Side-by-side example

`azapi_resource` (dynamic, today):

```hcl
resource "azapi_resource" "storage" {
  type      = "Microsoft.Storage/storageAccounts@2025-01-01"
  name      = var.name
  parent_id = azapi_resource.rg.id
  location  = var.location
  body = {
    kind = "StorageV2"
    sku  = { name = "Standard_LRS" }
    properties = {
      accessTier        = "Hot"
      minimumTlsVersion = "TLS1_2"
    }
  }
  response_export_values = ["properties.primaryEndpoints.blob"]
}
```

`azapi_storage_account` (azapin, static):

```hcl
resource "azapi_storage_account" "test" {
  name      = var.name
  parent_id = azapi_resource.rg.id
  location  = var.location
  kind      = "StorageV2"
  sku = { name = "Standard_LRS" }
  properties = {
    access_tier         = "Hot"
    minimum_tls_version = "TLS1_2"
  }
}
```

Both create the identical ARM resource. The difference is entirely in the authoring
and plan experience.

---

## Authoring & editor experience

| Feature | `azapi_resource` | `azapi_storage_account` (azapin) |
|---|---|---|
| Resource shape | one `body = { … }` blob | typed named attributes (`sku`, `properties`, …) |
| Property names | raw ARM **camelCase** (`accessTier`) | **snake_case** (`access_tier`) |
| Editor autocomplete | yes in VS Code via the Azure Terraform extension/LSP for `body`, but not from the provider schema itself | full, native — every attribute is in the provider schema |
| Plan-time type checking | yes — dynamic `body` validation in the provider (`schemaValidate`, default on); VS Code extension/LSP adds editor assistance | yes — native typed schema; wrong type/typo caught by tooling *and* `plan` |
| Reading computed outputs | `response_export_values` + `.output.properties.primaryEndpoints.blob` | direct ref: `properties.primary_endpoints.blob` |

## Plan & diff quality

> Both render **per-property** diffs. `body` is a structured dynamic object, so
> Terraform recurses into it and shows nested changes
> (`~ minimumTlsVersion = "TLS1_2" -> "TLS1_5"`), just like a typed nested attribute.
> The differences below are about *which* properties appear and how they're named —
> not granularity.

| Feature | `azapi_resource` | azapin |
|---|---|---|
| Diff rendering | per-property, nested under `body`, ARM **camelCase** | per-property, top-level typed attributes, **snake_case** |
| Computed / read-only values | not part of the resource; surfaced via `response_export_values` → `.output` (and only the paths you export) | first-class **typed Computed attributes** — visible in state, referenceable directly; round-trip without perpetual diff (UseStateForUnknown) |
| Defaults visible at plan | no — only what you write is planned (no default injection) | **yes** — verified defaults shown at plan (`minimum_tls_version = "TLS1_2"`, `public_network_access = "Enabled"`, `supports_https_traffic_only = true`, …) |

## Validation (catches errors before apply)

> Both validate against the embedded ARM/bicep types **at plan time**.
> `azapi_resource` does this via `schemaValidate` in `ValidateConfig` — **on by
> default** (`schema_validation_enabled = true`), traversing the dynamic `body`.
> azapin expresses the same constraints as a **native typed schema**, so the
> diagnostics, autocomplete, and `terraform validate` schema-awareness come from
> the provider schema itself rather than depending on the external VS Code
> extension/LSP used for `azapi_resource` body authoring.

| Feature | `azapi_resource` | azapin |
|---|---|---|
| Plan-time validation vs ARM schema | **yes** — `schemaValidate` vs embedded bicep (default on; toggle `schema_validation_enabled`) | **yes** — native schema validators + the same bicep-derived rules |
| Where errors point | into `body` JSON paths (e.g. `body.properties.accessTier`) | the specific typed attribute |
| Editor autocomplete / schema awareness | yes in VS Code via the Azure Terraform extension/LSP, but not native provider schema | yes — every attribute is in the provider schema |
| Enum / range / length / regex surfaced *in the schema* | no (checked at plan via traversal, not declared in schema) | yes — `OneOf`, ranges, length, regex on the attribute |
| Required-field checks | yes (body traversal) | per-attribute `Required` in schema |

## Lifecycle correctness (AzureRM-grade behavior)

| Feature | `azapi_resource` | azapin |
|---|---|---|
| ForceNew (which change forces replace) | not modeled per-property — body changes do an in-place PUT; an immutable property is only caught when ARM rejects the update at apply (our branch adds whole-`body` replace via azwise when a ForceNew property changes) | **per-attribute `RequiresReplace`** — the plan shows exactly which attribute forces replacement, before apply |
| Conditional ForceNew | not expressible as a plan modifier | **yes** — e.g. SKU **zone migration** (`Standard_LRS → Standard_ZRS`) forces replace; in-tier change doesn't |
| Read-only fields | user must know which ARM fields are read-only and not set them | marked `Computed`; stripped from the request automatically |
| Sensitive fields | manual `sensitive_body` | marked `Sensitive: true` in schema |
| Operation timeouts | generic defaults | per-resource defaults (create/read/update/delete) |
| Import | generic | typed import → state |

---

## Per-resource business logic (the strategic differentiator)

What makes AzureRM resources *valuable* isn't the typed schema — it's the
hand-written, per-resource intelligence layered on the API. azapin gives every
resource type a place to put that logic. `azapi_resource` structurally can't: it is
one generic engine with no per-resource extension point.

The static resource exposes typed **customization seams** (`internal/native/resource`):

- **Hooks** (data-driven): `BeforeCreate` / `AfterCreate` / `BeforeUpdate` /
  `AfterUpdate` / `BeforeRead` / `AfterRead` / `BeforeDelete`, plus `ValidateConfig`
  and `ModifyPlan`, registered per resource name. A hook can tweak the request body,
  post-process state, add cross-field validation, or call extra APIs.
- **Method override** (escape hatch): embed `*Base` and override any framework
  method for wholesale-custom behavior.
| Capability | `azapi_resource` | azapin |
|---|---|---|
| Per-resource extension point | none (one generic engine) | Hooks + method override |
| Cross-field validation | only generic body-vs-schema | arbitrary Go (e.g. "Premium tier ⇒ BlockBlobStorage kind") |
| Conditional ForceNew | not expressible | shipped — SKU zone migration |
| Field-dependent defaults / normalization | none | arbitrary (location/ID casing, derived values) |
| **Multi-API orchestration in one resource** | impossible — one PUT to one API | a hook can call extra APIs |
| ARM quirk / eventual-consistency workarounds | user works around them in HCL | encapsulated in the resource |
**Why the multi-API seam matters.** AzureRM's `azurerm_storage_account` bundles
`blob_properties` / `share_properties` / `queue_properties` — which are actually
*separate* ARM APIs (`BlobServices` / `FileServices` / `QueueServices.SetServiceProperties`).
That's why azwise deliberately excludes them from the storage account body. azapin's
default is **one resource per ARM type**, so today those ship as their own static
resources (e.g. `azapi_storage_account_blob_service`). Where AzureRM-style bundling is
genuinely worth it, an `AfterCreate` / `AfterUpdate` hook lets a single
`azapi_storage_account` make those extra calls and expose them as nested attributes —
a deliberate, hand-authored per-resource opt-in, not the default. `azapi_resource`
cannot do either. (See ADR-0004.)

**Cost framing (for leaders).** This is **opt-in and incremental**: the generated
schema + azwise overlay handle the long tail with *zero* custom code; you write a
hook only where a resource genuinely needs one. So you pay AzureRM's per-resource
logic cost only on the resources that warrant it — on top of mechanically generated
schemas — which is far cheaper than hand-authoring every resource from scratch.

---

## Under the hood (what powers the "intelligence")

- **Generated from Azure's bicep type graph** — the same type data the provider
  already embeds. The generator covers ~2,631 stable resource types; the storage
  account shipped first as the proof of concept.
- **azwise overlay** — hand-verified AzureRM-derived knowledge layered onto the
  generated schema: ForceNew rules, real default values, validation rules, and
  sensitive/computed classification. This is what makes the resource behave like
  AzureRM rather than a raw spec dump.

---

## Why both exist — `azapi_resource` still wins at

- **Day-zero coverage** — works for *any* ARM type / API version the moment Azure
  ships it; no codegen or provider release needed.
- **Preview APIs & polymorphic resources** — discriminated / edge resources azapin
  can't statically type.
- **Exact ARM names & escape hatch** — when the spec is wrong or you need raw control.

**Recommendation:** use the static `azapi_<service>_<resource>` for the common,
stable resources where you want the AzureRM-like experience; fall back to
`azapi_resource` for everything else.

---

## Honest status

- **Working & verified:** typed schema generation, framework schema-validity check,
  full CRUD applied against **live Azure**, and defaults / validation / ForceNew
  surfacing correctly in `terraform plan`. Storage account is the proven PoC.
- **In progress:** confirming full plan idempotency (empty re-plan) end-to-end, and
  generating/shipping resources beyond storage.
- **Known limits:** discriminated / polymorphic bodies and free-form maps with complex
  value types (e.g. user-assigned identities) aren't statically typed — they fall back
  to dynamic handling or `azapi_resource`. (String-valued maps like `tags` *are* fully
  typed.)

---

## Note for anyone testing

The AzureRM-knowledge features above (validation, ForceNew, defaults, sensitive
handling) are provided by the **azwise** subsystem, which is **not in the official
azapi provider** — it is our enhancement in this branch. For `azapi_resource` it can be
turned off with the `disable_resource_knowledge` provider feature. It does **not** apply
to azapin static resources: their knowledge is compiled into the typed schema at
generation time and cannot be un-baked at runtime — the escape hatch there is to use
`azapi_resource` instead. (See ADR-0005.)
