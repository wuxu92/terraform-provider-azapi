# Tester context — extend hook-contract test for the new `AfterDelete` hook

## Where
- File to EDIT (do not create new): `internal/native/resource/base_hook_contract_test.go`, `package resource`.
- This file is the AUTHORITY that pins the per-hook field-liveness matrix in `hooks.go`. A new hook `AfterDelete` was added; the contract test must now cover it in two places.

## Background — the hook just added
`Hooks` (in `internal/native/resource/hooks.go`) gained:
```go
AfterDelete func(*CrudCtx)
```
It fires in `Base.Delete` (`base.go`) AFTER the ARM DELETE succeeds (or 404s — NotFound is tolerated), and is BYPASSED on the Singleton reset path (which `return`s before it, same as `BeforeDelete`).

Liveness at `AfterDelete`: **State live; Plan, Body, Response null** — identical to `BeforeDelete`. (It reads the prior state of the just-deleted resource; there is no Response because DELETE returns nothing to map.)

## Change 1 — extend the `t.Run("delete", …)` subtest inside `TestHookContractFieldLiveness`
Current (around lines 248-271) registers only `BeforeDelete` and asserts its liveness. Extend it to ALSO register an `AfterDelete` recording hook, assert it ran, and assert its liveness. Mirror the existing `create`/`read` subtests that record two hooks (`beforeRan/afterRan`, `before/after` snapshots via `snapCrud`).

Concretely:
- Add `var afterRan bool` and `var after liveness`.
- In the `b.hooks = &Hooks{...}` literal, add `AfterDelete: func(hc *CrudCtx) { afterRan = true; after = snapCrud(hc) }` alongside the existing `BeforeDelete`.
- After the existing `beforeRan` check, add: if `!afterRan { t.Fatal("AfterDelete hook was never invoked") }`.
- Add: `assertLiveness(t, "AfterDelete", after, liveness{planLive: false, stateLive: true, bodyLive: false, responseLive: false})`.
- Keep the existing `BeforeDelete` assertion unchanged.

The existing `scriptClient` returns `http.StatusOK, "{}"` for the DELETE, so both hooks fire cleanly.

## Change 2 — pin that `AfterDelete` is bypassed on the Singleton reset path
`TestSingletonDeleteSkipsBeforeDelete` (around lines 279-303) proves the Singleton reset path skips `BeforeDelete`. Extend it to ALSO prove it skips `AfterDelete`:
- Add `var afterRan bool`.
- In the `b.hooks = &Hooks{...}` literal (which has `Singleton` + `BeforeDelete`), add `AfterDelete: func(hc *CrudCtx) { afterRan = true }`.
- After the existing `if beforeRan { t.Fatal(...) }`, add: `if afterRan { t.Fatal("AfterDelete ran on the Singleton reset path; the contract says it must be bypassed") }`.
- Update the test's doc comment to mention both `BeforeDelete` and `AfterDelete` are bypassed. Consider it fine to leave the function name as-is, or note in the comment that it now covers both delete hooks.

## Constraints
- Reuse existing helpers already in the file: `snapCrud`, `liveness`, `assertLiveness`, `scriptClient`, `resourceGroupReadState`, `resourceGroupDescriptor`, `rgArmID`. Do NOT invent new helpers.
- No production-code edits. No live Azure/TF_ACC/network. Pure stdlib `testing`.
- Do not touch other tests in the file.

## Verify
`go test ./internal/native/resource/... -run 'TestHookContractFieldLiveness|TestSingletonDeleteSkipsBeforeDelete' -count=1` passes; `gofmt -l` clean on the file. Report the passing output tail.
