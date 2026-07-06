# No static property map — resolve ARM names at runtime from the embedded type graph

Static resources do **not** carry a compiled per-resource table mapping snake_case Terraform attributes to ARM camelCase property names. Instead, at runtime the `Base` CRUD path loads the bicep `types.json` for the resource's pinned `ARMType@APIVersion` from the already-embedded schema index (`resource/loader.go`) and resolves names directly from that type graph (memoized per version).

We chose this over emitting a static map because a compiled table would duplicate ~215 property mappings across ~2,631 resource types inside a binary that already embeds the full bicep type set (~334 MB) — pure redundancy. The type graph is already in the binary; reuse it.

Consequence — a load-bearing invariant: **every registered descriptor's `ARMType@APIVersion` must exist in the embedded index.** Nothing enforces this at registration; a stale pin surfaces only as a runtime `no types.json location for X@Y` at `terraform apply`. A fast-lane test iterating `services.Registry` and asserting `typesJSONLocation` resolves for each descriptor is the required guard (agreed; see grilling follow-ups).
