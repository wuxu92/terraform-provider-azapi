# disable_resource_knowledge is azapi_resource-scoped; azapin cannot honor it

The `disable_resource_knowledge` provider feature turns off azwise for the dynamic `azapi_resource` (every azwise call there is gated on the flag). It does **not** apply to azapin static resources, and cannot meaningfully be made to.

Reason: for azapin, azwise is applied at **generation time** — `Sensitive`, `Computed`, `Required`, validators, and `RequiresReplace` are baked into the compiled `_gen.go` schema. A runtime flag cannot un-bake a compiled schema. The azapin `Base` also calls azwise at runtime (`StripComputedFields`, `TimeoutDefault`, hook `CheckForceNew`) unconditionally, and gating those would only cause harm (e.g. pushing computed fields into the PUT so ARM rejects it) for no benefit.

The honest escape hatch for "I don't want the knowledge layer" on a typed resource is therefore **use `azapi_resource` instead** — the dynamic surface where the flag does apply.

We record this because a future reader will reasonably expect the flag to disable a typed resource's ForceNew/validation and be surprised it doesn't. (Note: `azapin-vs-azapi.md` currently claims the flag makes azapin "behave like upstream"; that is false and must be corrected.)
