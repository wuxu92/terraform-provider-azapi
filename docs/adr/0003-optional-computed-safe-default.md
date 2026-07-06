# Optional+Computed is the mechanical default; azwise downgrades to plain Optional

For any body property that is neither `Required` nor `ReadOnly` and has no curated default, the generator emits `Optional + Computed` with `UseStateForUnknown` (the **safe default**). This prevents perpetual plan diffs on fields ARM populates server-side that the user never set — unavoidable when generating ~2,631 types where we usually cannot know which fields the server fills in.

The cost: `Optional+Computed` forfeits **unset-by-omission**. Once a value is in state, deleting it from HCL produces no diff — `Expand` skips null/unknown values, so the property is simply absent from the PUT and ARM keeps its prior value. The only way to clear such a field is an explicit zero value (if ARM honors it), or falling back to `azapi_resource`.

Where azwise has evidence a property is genuinely user-owned and clearable, it downgrades that property to plain `Optional`, restoring unset-by-omission. The limitation is documented per-resource in generated docs.

We chose this over a plain-`Optional` default because a wrong `Optional` on a server-populated field yields a perpetual diff (loud, breaks `terraform plan` idempotency), whereas a wrong `Optional+Computed` only costs unset-by-omission (recoverable via explicit value or azwise). Flipping `Computed` off later is a breaking schema change, so the conservative default is correct.
