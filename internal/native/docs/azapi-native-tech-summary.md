## A. Implemented features & key tech choice (9)

1. **Static typed schema generation from bicep `types.json`.**
   A type-graph walker resolves ARM `$ref` graphs into a Terraform
   plugin-framework schema. Key choices: latest-stable API version only, cycle
   breaking to a DAG, **description mining** (defaults/validators/format hints
   the formal spec omits), and conditional imports (pre-scan the graph so a
   generated file imports only what it uses).

2. **azwise operational-knowledge overlay.**
   Per-resource curated knowledge (ForceNew, real defaults, validation,
   sensitive/computed classification, timeouts) extracted from AzureRM and
   applied on top of the mechanical schema. This is what makes the output
   behave like a hand-authored resource. Key choice: azwise is the base for the
   override layer, not a runtime lookup.

3. **One generic runtime base — no per-resource Go struct** (ADR-0002).
   A single `Base` implements the full CRUD lifecycle for every generated
   resource; ARM property names are resolved at runtime from the embedded type
   graph, never duplicated per resource.

4. **Generic bidirectional mapper (Terraform ⇄ ARM JSON).**
   `Expand` (typed config → ARM payload) and `Flatten`/`FlattenInto` (ARM
   response → typed state) walk the shared type graph — no per-resource
   marshaling code. Handles idempotency concerns: sensitive/write-only
   retention, semantic location equality, case-preserving open-map keys.

5. **Two customization seams: generation-time customizers + runtime hooks.**
   Customizers mutate the type graph before emission (validators, defaults,
   ForceNew, envelope) so the shipped schema is always final; runtime hooks
   (`Before/After Create/Read/Update/Delete`, Singleton, `ModifyPlan`,
   `ValidateConfig`) customize behavior only. Key choice: schema is frozen at
   generation time; the runtime never rewrites it (ADR-0007).

6. **Relational cross-property constraints.**
   `ConflictsWith / RequiredWith / ExactlyOneOf / AtLeastOneOf` lowered from
   azwise into the generated descriptor and enforced at runtime as framework
   `ConfigValidators` over nested snake_case paths.

7. **Operational envelope synthesis.**
   `name`, parent reference, `id`, and per-operation `timeouts` are synthesized
   outside the bicep body (bicep does not model resource naming/placement) and
   seeded from the ARM type + writable scope.

8. **Data source auto-generation** (shipped).
   Every registered resource gets a read-only data source for free: the generic
   `DataSource` converts the resource schema (name + parent Required, all else
   Computed) and reuses the resource read path.

9. **Self-verification gate.**
   Generation runs flag-invariant checks plus a schema↔bicep **parity
   validator** (extra/type-mismatch fatal, missing = warning). A resource only
   graduates the allowlist after invariants + a live-Azure acceptance test + an
   authored azwise overlay (ADR-0006). Every shipped resource is a permanent
   schema contract, so correctness gates coverage.

---

## B. Key future features & possible solutions (6)

10. **Discriminated / polymorphic type support** (biggest typed-coverage gap, undergoing...).
    Today every `DiscriminatedObjectType` collapses to `types.Dynamic`; **632
    resources whose root body is discriminated are skipped entirely** and fall
    back to `azapi_resource`. Recommended solution: **nested per-variant blocks**
    (one Optional `SingleNestedAttribute` per variant + `ExactlyOneOf`,
    discriminator derived), reusing existing nested-object + relational
    machinery, with a variant-count fallback to dynamic for pathological wide
    unions. Full analysis in `discriminator-report.md`.

11. **Live-API schema verification harness** (TODO #16).
    Bicep/azwise are lossy vs real ARM behavior (ranges, enum completeness,
    MaxItems, immutability, real defaults, write-but-not-honored fields).
    Solution: probe the live API with boundary values, diff PUT→GET, classify
    each outcome (Confirmed / Refuted / Coerced / Inconclusive) into suggested
    azwise/customizer edits. One core, two adapters — raw REST engine +
    typed-TF regression path (ADR-0008).

12. **`azapi_resource` → native migration tooling** (TODO #9, #5).
    Read `type/name/parent_id/body/response_export_values`, map body JSON paths
    to native attributes, emit new HCL. Hardest part is state migration and a
    fallback for polymorphic/dynamic fields; AzureRM→native is a larger superset
    needing name mapping + behavior-diff reports.

13. **Lifecycle tooling: breaking-change detector + API-version upgrade + docs**
    (TODO #6, #8, #2). Compare old vs regenerated schema models (removed attrs,
    flag/type/validator/default/ForceNew changes) to drive deprecation windows
    and framework state upgraders; automate version-bump regeneration + azwise
    path-survival checks; generate AzureRM-style per-resource docs from
    descriptors + azwise + config-builder examples.

14. **Terraform resource identity + list resources** (TODO #12, #13).
    Implement `ResourceWithIdentity` on `Base` (immutable ARM identity, import
    by identity) and typed list resources with paging/filtering, reusing the new
    data-source read/flatten machinery.

15. **Cross-resource operation locking** (TODO #3).
    Scoped, keyed lock manager (by ARM ID / parent ID) for operations ARM
    serializes poorly — subnets, app settings, role assignments, network rules.
    Avoid global locks.

---

## C. Non-goals / hard limits (4)

16. **Not a replacement for `azapi_resource`.** It remains the day-zero,
    any-type/any-version, preview, and escape-hatch path. azapin coexists;
    users choose per resource.

17. **No per-resource Go model structs, and no static ARM-name property map**
    (ADR-0002). Everything is generic + runtime-resolved from the embedded type
    graph. This bounds binary/code growth but means all behavior flows through
    one generic engine.

18. **No AzureRM property-name parity.** azapin uses **mechanical snake_case**
    from the spec — predictable, not curated. Migration tooling must bridge the
    naming gap; parity is explicitly out of scope.

19. **Latest stable API version only, one resource per ARM type** (ADR-0001,
    ADR-0004). No user-settable `api_version`; preview versions →
    `azapi_resource`. Multi-version support is deliberately deferred (multiplies
    schemas, docs, tests, and state transitions). Sub-APIs are modeled as their
    own resources, not nested.
