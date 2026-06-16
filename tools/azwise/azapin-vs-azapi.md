# azapin (static resources) vs `azapi_resource` (dynamic)

**Audience:** teammates and leadership.
**One-liner:** azapin turns a single freeform `azapi_resource` (raw ARM JSON) into a
typed, per-resource Terraform resource generated from Azure's own schema — with
autocomplete, real validation, correct diffs, and AzureRM-grade operational
knowledge — while `azapi_resource` stays the day-zero escape hatch.

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
| Editor autocomplete | none (`body` is opaque) | full — every attribute is in the schema |
| Plan-time type checking | none | yes (wrong type / typo caught at `plan`) |
| Reading computed outputs | `response_export_values` + `.output.properties.primaryEndpoints.blob` | direct ref: `properties.primary_endpoints.blob` |

## Plan & diff quality

| Feature | `azapi_resource` | azapin |
|---|---|---|
| Plan diff granularity | whole `body` shown as one change | per-attribute diff (`access_tier: "Hot" → "Cool"`) |
| Computed values | mostly `(known after apply)` blob | server values round-trip into state; **no perpetual diff** (UseStateForUnknown) |
| Defaults visible at plan | no — you must set everything | **yes** — verified defaults shown in plan (`minimum_tls_version = "TLS1_2"`, `public_network_access = "Enabled"`, `supports_https_traffic_only = true`, …) |

## Validation (catches errors before apply)

| Feature | `azapi_resource` | azapin |
|---|---|---|
| Enum validation | optional runtime (`schema_validation_enabled`) | schema-level `OneOf` (e.g. `access_tier ∈ {Hot,Cool,Cold,Premium,Smart}`) |
| Numeric ranges / string length / regex | none | generated from spec + descriptions |
| ARM resource-ID format | none | regex validator on ID-shaped fields |
| Required-field checks | runtime, generic | per-attribute `Required` in schema |

## Lifecycle correctness (AzureRM-grade behavior)

| Feature | `azapi_resource` | azapin |
|---|---|---|
| ForceNew | generic "body changed → replace" | **per-attribute `RequiresReplace`** (only the right fields force replace) |
| Conditional ForceNew | not expressible | **yes** — e.g. SKU **zone migration** (`Standard_LRS → Standard_ZRS`) forces replace; in-tier change doesn't |
| Read-only fields | user must know not to set them | marked `Computed`; stripped from the request automatically |
| Sensitive fields | manual `sensitive_body` | marked `Sensitive: true` in schema |
| Operation timeouts | generic defaults | per-resource defaults (create/read/update/delete) |
| Import | generic | typed import → state |

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
- **Known limits:** free-form map fields (e.g. `tags` contents, user-assigned
  identities) and discriminated / polymorphic bodies aren't fully typed yet — those
  fall back to dynamic handling or `azapi_resource`.

---

## Note for anyone testing

The AzureRM-knowledge features above (validation, ForceNew, defaults, sensitive
handling) are provided by the **azwise** subsystem, which is **not in the official
azapi provider** — it is our enhancement in this branch. It can be turned off with
the `disable_resource_knowledge` provider feature, which makes the build behave like
upstream for those paths.
