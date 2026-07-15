# AzAPI Native TODO Report

Grounding: current native design/spec docs, acceptance framework, provider registration, and Terraform Plugin Framework resource identity docs. Workload estimates are [INFERENCE].

## 1. Improve acceptance tests: BDD, dependency management, DI, customize config

Current Ginkgo framework has `Workspace`, `Scope`, `ResourceFor`, `Stage`, drift checks, import verify, and config builders. Next work: formal dependency graph, reusable fixtures, DI for locations/quotas/provider config, negative cases, and per-resource scenario composition.  
**Workload:** [INFERENCE] 1-2 weeks for framework polish; ongoing per-resource adoption.

## 2. Native resource documentation generator

Generate user-facing docs per native resource, similar to AzureRM provider docs. Source schema from generated descriptors plus bicep/swagger descriptions, azwise defaults/ForceNew/validation, examples from config builders, import format from ARM type, and nested attribute tables. Publish generated Markdown with stable anchors and regeneration diff checks.  
**Workload:** [INFERENCE] 2-4 weeks for generator and template; per-resource examples ongoing.

## 3. Lockers between resources/API operations

Native resources currently share the generic runtime base and ARM client; no native operation locking was found. Need keyed locks by ARM ID or parent ID for operations ARM serializes poorly: subnet, app settings, role assignments, network rules. Avoid global locks; implement scoped provider-level lock manager.  
**Workload:** [INFERENCE] 1 week base lock manager; 1-3 days per resource rule family.

## 4. Automate data source support — DONE

Shipped in `resource/datasource.go` + `resource/schema_convert.go`: `Provider.DataSources()`
now appends `nativeresource.NewDataSource(name)` for every `services.Registry` entry,
mirroring `Resources()`. The generic `DataSource` converts the generated resource schema
into a read-only data-source schema at runtime (`name` + parent as Required inputs, all
else Computed) and reuses the resource read path (compose ARM id → GET → mapper flatten).
Lookup identity is `name` + `<parent>_id` (no per-resource `resource_id` alt-path yet).
See RESOURCE.md "Data Source". Remaining (optional): a `resource_id` lookup alternate.

## 5. Static resource convert from AzureRM to AzAPI native

This is larger than schema generation. Need mapping from AzureRM resource names/properties to native AzAPI names, state migration guidance, HCL rewrite tooling, and behavior-diff reports. Azwise extraction helps but does not preserve AzureRM naming parity by design.  
**Workload:** [INFERENCE] 4-8 weeks for first useful converter; months for broad coverage.

## 6. Generated schema breaking-change detector and mitigator

Add artifact comparison between old and regenerated schema models: removed attributes, flag changes, type changes, validator tightening, default changes, ForceNew changes. Mitigation: deprecation windows, state upgraders, aliases only when safe, and release-blocking reports.  
**Workload:** [INFERENCE] 2-3 weeks for detector; mitigation rules ongoing.

## 7. Preview API version resource support

Current design explicitly uses latest stable; preview falls back to `azapi_resource`. Supporting preview requires opt-in target selection, naming/version policy, release stability rules, and docs warning that preview schemas can churn. Avoid default preview resources in normal registry.  
**Workload:** [INFERENCE] 1-2 weeks for opt-in generation; higher maintenance risk.

## 8. API version upgrade tools

Developer guide has a manual upgrade checklist. Automate version selection, regenerate target, diff schema, validate azwise path survival, highlight removed/renamed fields, run focused tests, and generate upgrade notes. This should depend on the breaking-change detector.  
**Workload:** [INFERENCE] 2-4 weeks; less if schema diff infrastructure lands first.

## 9. Dynamic `azapi_resource` to native resource migration

Build migration tooling that reads `type`, `name`, `parent_id`, `body`, and `response_export_values`, maps body JSON paths to native attributes, and emits new HCL. Need fallback for polymorphic/dynamic fields and API version mismatches. State migration remains hardest.  
**Workload:** [INFERENCE] 3-6 weeks for MVP; broad reliability needs real-world corpus testing.

## 10. Cross-property/block constraints and custom diff management — PARTIALLY DONE

The declarative constraint engine **shipped**. Azwise `ConflictsWith / RequiredWith / ExactlyOneOf / AtLeastOneOf` rules are lowered at generation time (`typegraph/azwise_overlay.go`) into `services.RelationalConstraint` on the descriptor (`services/registry.go`), emitted by the generator (`generator/emitter.go`), and enforced at runtime as reusable framework `ConfigValidators` over nested snake_case paths (`resource/configvalidators.go`). Mutual-exclusion members also drop `UseStateForUnknown` via `SuppressStateReuse` so an omitted side clears instead of pinning stale state. Live examples: `azapi_virtual_network` `address_space` ↔ `ipam_pool_prefix_allocations` ExactlyOneOf; `azapi_storage_account` CMK RequiredWith.
Remaining: a **declarative diff suppress/normalization policy for server-mutated fields** — today handled ad hoc via per-property customizer flags (`NonNullStateForUnknown`, `DefaultEmptyList`, `UseSet`) and runtime hooks, not a general rule surface.  
**Workload:** [INFERENCE] constraint engine done; ~1-2 weeks for a declarative normalization policy; per-resource rules ongoing.

## 11. Deprecation, breaking changes, and state migration

Need formal lifecycle: mark deprecated in generated schema, emit diagnostics, keep old attributes until major/provider-defined cutoff, and write framework state upgraders for renames/type moves. Resource deprecation also needs registry metadata and docs generation. Depends on schema breaking detector.  
**Workload:** [INFERENCE] 3-5 weeks for framework; each migration varies from hours to days.

## 12. Terraform resource identity support

This is Terraform Plugin Framework resource identity, not Azure managed identity blocks. Implement `ResourceWithIdentity` on the native `Base`, emit an `IdentitySchema` for immutable ARM identity data, probably `id` first, then set identity during Create/Read/Update and support import by identity while keeping ID import.  
**Workload:** [INFERENCE] 1-2 weeks for base wrapper; more if identity schema includes decomposed ARM ID parts.

## 13. Resource list support

Provider already has protocol `ListResources()` with `AzapiResourceList`; native has no generated list resource registry. Need typed list schemas, paging, filter inputs, output shape decisions, and mapping list responses to collections. Also clarify Terraform list-resource UX versus data sources.  
**Workload:** [INFERENCE] 2-4 weeks. The native data-source base (item 4) now exists, so the read/flatten and schema-reuse machinery can be reused.

## 14. Old API version or multiple API version support

Current design is one native resource per ARM type at latest stable. Multiple versions multiply schemas, docs, tests, state transitions, and azwise matching. Low priority is correct. If needed, prefer explicit versioned resource names or an advanced `api_version` only when schema-compatible.  
**Workload:** [INFERENCE] 4-8 weeks plus high ongoing maintenance.

## 15. Move azwise out of azapi provider repo

Azwise is currently compiled into provider behavior and generator overlay. Extracting it needs module boundary, versioning, generated artifacts or Go API, validation access to Azure SDK types, and synchronized provider/generator releases. Benefit: reuse by tools and independent knowledge lifecycle.  
**Workload:** [INFERENCE] 3-6 weeks initial extraction; release/process cost continues.

## 16. Live-API schema verification harness

Generated native schema derives from lossy sources (bicep/swagger + azwise/AzureRM); neither faithfully mirrors ARM API behavior for validation ranges, enum completeness, `MaxItems`, `Required`, immutability/ForceNew, real defaults, or write-accepted-but-not-honored properties. Design a mechanism to verify per-property knowledge against the live API by probing with boundary values and diffing PUT→GET, classifying each outcome (Confirmed / Refuted / Coerced / Inconclusive) into suggested azwise/customizer edits. Primary engine = raw `ResourceClient` REST (bypasses the typed schema that would otherwise filter the signal); typed-resource TF path (`nativeacc.Workspace`, `-debuggable`) reserved for regression and interactive debugging. See `SCHEMA_VERIFICATION.md`. Depends conceptually on #6 (schema diff) and #10 (constraint engine).  
**Workload:** [INFERENCE] 1 week REST PoC on one resource; 2-4 weeks planner+classifier+reporter; batching/governor + CI cadence ongoing.
