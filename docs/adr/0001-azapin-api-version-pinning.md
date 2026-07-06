# Azapin resources pin one ARM API version, invisible to users

Each azapin-generated static resource bakes in exactly one ARM API version (the latest stable at generation time). There is no user-settable `api_version` attribute: the typed schema, its validators, and the runtime camelCase name resolution all correspond to that single pinned version.

We chose this over a user-settable version because the core guarantee of a static resource is that its typed schema faithfully mirrors one API version's body — a runtime override would let a user pick a version whose real body diverges from the emitted schema, silently breaking validation and diffs. `azapi_resource` remains the escape hatch for arbitrary versions.

Advancing the pin is therefore a governed regeneration event (gated by the breaking-change detector, TODO #6) surfaced via release notes and, where needed, state upgraders — not a runtime knob.
