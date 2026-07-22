# disable_resource_knowledge and azapi_resource runtime knowledge were removed (temporary experiment)

**Superseded.** The `disable_resource_knowledge` provider feature and the azwise-derived
runtime checks it gated on the dynamic `azapi_resource` (naming/property validation,
body-property `CheckForceNew`, computed-field stripping, per-operation timeout defaults)
were a temporary experiment and have been removed.

Reason: azwise and its generated knowledge are moving to a **separate repository**, used
only at **generation time** by the azapin generator overlay. Wiring the dynamic
`azapi_resource` to the azwise runtime registry created a runtime dependency that blocks
that extraction, for a knowledge surface the static (azapin) resources already bake into
their compiled `_gen.go` schema. The dynamic resource now relies on ARM itself as the
validator and on plain per-operation timeout defaults.

Consequences:
- The `disable_resource_knowledge` provider argument and its `ARM_DISABLE_RESOURCE_KNOWLEDGE`
  environment variable no longer exist.
- `azapi_resource` no longer imports `github.com/wuxu92/azwise`.
- The generated native service packages (`internal/native/services/*`) no longer import
  azwise: the two value-conditional `ModifyPlan` hooks (storage, documentdb) inline their
  ForceNew predicates instead of calling `azwise.CheckForceNew`.
- The azapin runtime `Base` still calls `azwise.StripComputedFields` and the generator
  overlay (`typegraph/azwise_overlay.go`) still applies azwise; decoupling those from the
  provider binary is the remaining step before azwise can physically leave the repo.

The historical scope note this ADR originally recorded (the flag was `azapi_resource`-scoped
and could not un-bake a compiled azapin schema) is now moot: the flag is gone entirely.
