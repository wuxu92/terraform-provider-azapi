## A. Implemented features & key tech choice (11)

1. **Static typed schema generation from bicep `types.json`.**
   A type-graph walker resolves ARM `$ref` graphs into a Terraform
   plugin-framework schema. Key choices: latest-stable API version only, cycle
   breaking to a DAG, **description mining** (defaults/validators/format hints
   the formal spec omits), and conditional imports (pre-scan the graph so a
   generated file imports only what it uses).
   **Why:** a typed, plan-time-validated schema gives practitioners
   autocomplete, required/enum/range checks, and diffable attributes the opaque
   `azapi_resource` body cannot — the payoff that justifies generating and
   shipping a per-resource schema at all.

2. **azwise operational-knowledge overlay.**
   Per-resource curated knowledge (ForceNew, real defaults, validation,
   sensitive/computed classification, timeouts) extracted from AzureRM and
   applied on top of the mechanical schema. This is what makes the output
   behave like a hand-authored resource. Key choice: azwise is the base for the
   override layer, not a runtime lookup.
   **Why:** bicep/swagger is lossy — no real ForceNew, defaults, validation
   ranges, or sensitive/computed truth — while terraform-provider-azurerm
   already encodes years of curated, live-tested operational knowledge; the
   overlay reuses that instead of re-deriving it, and applying it at generation
   time (not runtime) keeps the shipped schema final.

3. **One generic runtime base — no per-resource Go struct** (ADR-0002).
   A single `Base` implements the full CRUD lifecycle for every generated
   resource; ARM property names are resolved at runtime from the embedded type
   graph, never duplicated per resource.
   **Why:** emitting a per-resource Go struct + CRUD implementation for the
   hundreds-to-thousands of ARM types would explode binary size and multiply
   lifecycle bugs; one generic base keeps code growth flat and fixes every
   resource at once.

4. **Generic bidirectional mapper (Terraform ⇄ ARM JSON).**
   `Expand` (typed config → ARM payload) and `Flatten`/`FlattenInto` (ARM
   response → typed state) walk the shared type graph — no per-resource
   marshaling code. Handles idempotency concerns: sensitive/write-only
   retention, semantic location equality, case-preserving open-map keys.
   **Why:** the idempotency-critical concerns (sensitive/write-only retention,
   semantic location equality, open-map key casing) are subtle and identical
   across resources — implementing them once in a graph-walking mapper is
   correct by construction, where N generated per-resource marshalers would each
   be a place to get them wrong.

5. **Two customization seams: generation-time customizers + runtime hooks.**
   Customizers mutate the type graph before emission (validators, defaults,
   ForceNew, envelope) so the shipped schema is always final; runtime hooks
   (`Before/After Create/Read/Update/Delete`, Singleton, `ModifyPlan`,
   `ValidateConfig`) customize behavior only. Key choice: schema is frozen at
   generation time; the runtime never rewrites it (ADR-0007).
   **Why (two reasons):** (1) what the developer reads in `<name>_gen.go` is
   exactly what ships to the customer, so debugging an issue never means merging
   a patch or overlay over the schema in your head. (2) The schema is
   regenerated on every swagger re-vendor and API-version bump — a hand-edited
   `_gen.go` would be silently overwritten, and a separate patch layer's
   compatibility with the newly-generated schema cannot be known — so all
   customization moves *into* the generator ahead of emission rather than onto
   the output.

6. **Relational cross-property constraints.**
   `ConflictsWith / RequiredWith / ExactlyOneOf / AtLeastOneOf` lowered from
   azwise into the generated descriptor and enforced at runtime as framework
   `ConfigValidators` over nested snake_case paths.
   **Why:** the cross-field rules are declared once in the knowledge layer and
   enforced by framework-native `ConfigValidators`, so they survive regeneration
   and need no hand-written per-resource validation Go.

7. **Operational envelope synthesis.**
   `name`, parent reference, `id`, and per-operation `timeouts` are synthesized
   outside the bicep body (bicep does not model resource naming/placement) and
   seeded from the ARM type + writable scope.

8. **AzureRM-authoritative resource naming.**
   The `azapi_*` resource noun for an ARM type derives from terraform-provider-
   azurerm's curated names, not a raw spec derivation, so a generated resource
   carries the name practitioners already know
   (`Microsoft.DocumentDB/databaseAccounts` → `azapi_cosmosdb_account`, not the
   mechanical `azapi_document_database_account`). An offline extractor
   (`naming/cmd/azurerm-armtypes`) traces every azurerm resource's Create ID
   constructor → go-azure-sdk `fmtString` → ARM type, emitting a generated table
   (`azurerm_reference_gen.go`) that `ResourceName` consults first, swapping the
   `azurerm_` prefix for `azapi_`; types absent from the table fall back to the
   mechanical service+segment rule. Ambiguous ARM types (several azurerm
   resources share one type) are auto-resolved by two heuristics — **main
   resource** (one member is a whole-word prefix of the rest, e.g.
   `.../service` → `api_management`) and **discriminated base** (members are a
   common prefix + a distinguishing suffix, e.g. `.../identityProviders` →
   `api_management_identity_provider`) — guarded so two distinct ARM types can
   never emit the same noun (collision with an existing name, or a base contested
   by multiple groups, drops the group back to the mechanical rule). Key choice:
   the generated table is **machine-owned** (regenerated, never hand-edited);
   imperfect collapses, extractor-unresolvable scope/data-plane IDs, and the
   canonical noun for still-ambiguous types are patched through override maps in
   the naming package (`azurermReferenceOverrides`, `ambiguousResourceNames`)
   that win at lookup, keeping already-generated resource names stable across
   regeneration.

9. **Data source auto-generation** (shipped).
   Every registered resource gets a read-only data source for free: the generic
   `DataSource` converts the resource schema (name + parent Required, all else
   Computed) and reuses the resource read path.
   **Why:** a read-only data source is fully derivable from the resource schema
   plus the read path, so hand-authoring one per resource would be pure
   duplication that drifts from the resource; deriving it gives every resource
   free, always-consistent data-source coverage.

10. **Self-verification gate.**
    Generation runs flag-invariant checks plus a schema↔bicep **parity
    validator** (extra/type-mismatch fatal, missing = warning). A resource only
    graduates the allowlist after invariants + a live-Azure acceptance test + an
    authored azwise overlay (ADR-0006). Every shipped resource is a permanent
    schema contract, so correctness gates coverage.

11. **Acceptance config-builder authoring agent** (ADR-0009).
    The Basic/Complete/Complete_update acceptance config builders
    (`<resource>_config.go`, ~200-370 LOC/resource) are the largest zero-tooled
    hand-written surface — the live-Azure test scaffolding the graduation gate
    (feature 10) requires. `skill://azapi-acceptance-author` drafts them in one
    shot by translating AzureRM's own proven `basic`/`complete`/`update`
    acceptance configs into azapin's ARM-body shape (snake_case under
    `properties`, envelope extraction, composite-field joins), reusing the azwise
    automap engine. Key choices: the agent **drafts but never runs** acceptance
    (live Azure is minutes/apply and human-gated), so **Azure is the oracle** —
    a property that cannot be confidently translated gets a best-guess value plus
    an inline `TODO(acceptance-author)` marker, never a silent omission (a wrong
    guess costs one live-apply round-trip Azure names precisely; omitting loses
    coverage). The generated schema is the typed allowlist (`go build` enforces
    every drafted attribute exists), and the native-dependency wall fails loud on
    a missing native type rather than emitting a foreign block.

---

## B. Key future features & possible solutions (6)

12. **Discriminated / polymorphic type support** (biggest typed-coverage gap, undergoing...).
    Today every `DiscriminatedObjectType` collapses to `types.Dynamic`; **632
    resources whose root body is discriminated are skipped entirely** and fall
    back to `azapi_resource`. Recommended solution: **nested per-variant blocks**
    (one Optional `SingleNestedAttribute` per variant + `ExactlyOneOf`,
    discriminator derived), reusing existing nested-object + relational
    machinery, with a variant-count fallback to dynamic for pathological wide
    unions. Full analysis in `discriminator-report.md`.

13. **Live-API schema verification harness** (TODO #16).
    Bicep/azwise are lossy vs real ARM behavior (ranges, enum completeness,
    MaxItems, immutability, real defaults, write-but-not-honored fields).
    Solution: probe the live API with boundary values, diff PUT→GET, classify
    each outcome (Confirmed / Refuted / Coerced / Inconclusive) into suggested
    azwise/customizer edits. One core, two adapters — raw REST engine +
    typed-TF regression path (ADR-0008).

14. **`azapi_resource` → native migration tooling** (TODO #9, #5).
    Read `type/name/parent_id/body/response_export_values`, map body JSON paths
    to native attributes, emit new HCL. Hardest part is state migration and a
    fallback for polymorphic/dynamic fields; AzureRM→native is a larger superset
    needing name mapping + behavior-diff reports.

15. **Lifecycle tooling: breaking-change detector + API-version upgrade + docs**
    (TODO #6, #8, #2). Compare old vs regenerated schema models (removed attrs,
    flag/type/validator/default/ForceNew changes) to drive deprecation windows
    and framework state upgraders; automate version-bump regeneration + azwise
    path-survival checks; generate AzureRM-style per-resource docs from
    descriptors + azwise + config-builder examples.

16. **Terraform resource identity + list resources** (TODO #12, #13).
    Implement `ResourceWithIdentity` on `Base` (immutable ARM identity, import
    by identity) and typed list resources with paging/filtering, reusing the new
    data-source read/flatten machinery.

17. **Cross-resource operation locking** (TODO #3).
    Scoped, keyed lock manager (by ARM ID / parent ID) for operations ARM
    serializes poorly — subnets, app settings, role assignments, network rules.
    Avoid global locks.

---

## C. Non-goals / hard limits (4)

18. **Not a replacement for `azapi_resource`.** It remains the day-zero,
    any-type/any-version, preview, and escape-hatch path. azapin coexists;
    users choose per resource.

19. **No per-resource Go model structs, and no static ARM-name property map**
    (ADR-0002). Everything is generic + runtime-resolved from the embedded type
    graph. This bounds binary/code growth but means all behavior flows through
    one generic engine.

20. **No AzureRM property-name parity.** Resource-type nouns are
    AzureRM-authoritative (feature 8), but **property** names within a resource
    stay **mechanical snake_case** from the spec — predictable, not curated.
    Migration tooling must bridge the property-naming gap; property parity is
    explicitly out of scope.

21. **Latest stable API version only, one resource per ARM type** (ADR-0001,
    ADR-0004). No user-settable `api_version`; preview versions →
    `azapi_resource`. Multi-version support is deliberately deferred (multiplies
    schemas, docs, tests, and state transitions). Sub-APIs are modeled as their
    own resources, not nested.
