# PRD: Surface the hook-author contract at the `Hooks`/`CrudCtx` seam

Status: ready-for-agent

_Source: architecture review of `internal/native` (2026-07-06), candidate 4 — "Move the hook-author contract into the seam". The dead-field half of that candidate (wiring `BeforeRead` into `Base.Read`) is already implemented and committed (`d6f1b2d01`); this PRD covers the remaining half._

## Problem Statement

A resource author extends a static resource by registering a Hook. The `Hooks`
and `CrudCtx` structs advertise a small, flat surface — a handful of callback
fields and an eight-field context object — but the rules that actually govern
writing a correct hook are not visible there. To learn them, the author has to
read `Base`'s orchestration:

- **Field liveness differs per hook point.** `CrudCtx` exposes `Plan`, `State`,
  `Body`, and `Response`, but only a subset is populated at any given hook. A
  `Before*` create/update hook sees `Body` mid-composition and no `Response`; an
  `After*` hook sees `Response` but must not expect a further `Body` mutation to
  reach ARM; a read/delete hook sees `State` but no `Plan`. Reading a field that
  is null at that hook point is a silent bug the struct does nothing to prevent
  or document.
- **Ordering relative to the base's own work is implicit.** Whether a hook runs
  before or after body composition, the pre-create existence check, the ARM
  GET/PUT, and computed-field stripping is decided entirely inside `Base` and
  invisible at the seam.
- **`Singleton` suppresses other hooks.** Marking a resource `Singleton` changes
  the create and delete paths (create skips the existence check; delete resets
  instead of issuing DELETE, and `BeforeDelete` does not run on that path). None
  of this is stated where `Singleton` is declared.
- **The docs disagree with each other and with the code.** The just-fixed
  `BeforeRead` case is the symptom: one doc called it a permanent no-op while
  another listed it as a live step, and neither matched `Base`. Nothing keeps the
  documented contract and the actual orchestration from drifting apart again.

The result: the hook interface looks shallow but is deep, and the depth is
learned by reading the implementation — exactly the knowledge burden a good seam
removes.

## Solution

Make the hook-author contract legible **at the seam** and **guarded against
drift**, without changing any runtime behavior:

1. **State the per-hook-point contract where the author reads it** — on the
   `Hooks` fields and the `CrudCtx` fields. For each hook point, declare which
   `CrudCtx` fields are live (populated and meaningful), which are null, whether
   a `Body` mutation is still sent to ARM, and where the hook sits in the base's
   ordered lifecycle (relative to body composition, existence check, GET/PUT,
   strip). State the `Singleton` suppression rules where `Singleton` is declared.

2. **Guard the contract with a test that fails when `Base` drifts from it.** A
   unit test drives each `Base` CRUD entry point with a recording hook and
   asserts the field-liveness the contract promises — so a future change to the
   orchestration that stops populating a documented-live field, or starts
   populating a documented-null one, breaks the build instead of shipping a
   silent contract violation.

The author then writes a correct hook from the seam alone; the maintainer gets a
single source of truth that the test keeps honest.

## User Stories

1. As a resource author, I want each `Hooks` field to tell me at what point in
   the CRUD lifecycle it fires, so that I know what has and hasn't happened when
   my callback runs.
2. As a resource author, I want each hook point to tell me which `CrudCtx` fields
   are populated when it runs, so that I don't read a `Plan`, `State`, `Body`, or
   `Response` that is null at that point.
3. As a resource author writing a `Before*` mutation hook, I want to know whether
   mutating `Body` still reaches ARM, so that I mutate in a hook that is actually
   upstream of the send.
4. As a resource author writing an `After*` hook, I want to know that `Response`
   is the post-GET ARM body and that later state mapping consumes it, so that I
   massage the right object.
5. As a resource author, I want the `BeforeRead` contract to state that it fires
   before the read GET with `State` live and `Response` null, so that I use it
   for preflight and never expect the response.
6. As a resource author, I want the `BeforeDelete` contract to state its `State`
   liveness and that it does not run on the `Singleton` reset path, so that I
   don't rely on it for a singleton default.
7. As a resource author marking a resource `Singleton`, I want the suppression
   and path-change rules stated where `Singleton` is declared, so that I
   understand which hooks still run and which the singleton path bypasses.
8. As a resource author, I want the hook-point ordering relative to the
   pre-create existence check and computed-field stripping stated at the seam, so
   that I place logic that depends on that ordering correctly.
9. As a maintainer, I want a single documented contract for hook field liveness
   and ordering, so that the two developer docs stop contradicting each other and
   the code.
10. As a maintainer, I want a test that fails when `Base` populates a hook's
    `CrudCtx` differently than the contract states, so that contract drift is
    caught at build time rather than by a confused resource author.
11. As a maintainer, I want the contract test to exercise every hook point
    `Base` invokes, so that no hook point's liveness is left unguarded.
12. As a maintainer, I want this to change zero runtime behavior, so that the
    change is a pure clarity/safety improvement reviewable without acceptance
    runs.
13. As a resource author, I want the developer guide's hook table and the read
    lifecycle to match the seam contract exactly, so that whichever surface I
    read first tells me the same thing.
14. As a future contributor adding a new hook point, I want the contract and its
    guarding test to make the expectation obvious, so that I document and guard
    the new point the same way.

## Implementation Decisions

- **Modules touched.** `internal/native/resource` only — the `Hooks`/`CrudCtx`
  declarations (documentation of the contract) and a new/extended white-box test
  in the same package. The developer-facing docs under `internal/native`
  (`DEVELOPER_GUIDE.md`, `DEVELOPER_SPEC.md`, `RESOURCE.md`) are reconciled to the
  same contract. `CONTEXT.md`'s **Hook** glossary entry is updated to drop the
  "`BeforeRead` … not yet invoked" caveat (now wired) and point at the seam
  contract as the authority.
- **The contract is per-hook-point field liveness + ordering + suppression.** For
  each of `BeforeCreate`/`AfterCreate`/`BeforeUpdate`/`AfterUpdate`/`BeforeRead`/
  `AfterRead`/`BeforeDelete`, the live/null status of `Plan`, `State`, `Body`,
  `Response` (with `Ctx`/`Client`/`ID`/`Diags` always live), whether a `Body`
  mutation is sent, and the hook's position in the base's ordered lifecycle. The
  exact matrix is derived from `Base`'s current orchestration — this PRD does not
  restate it line by line; the implementing agent reads `Base` as the ground
  truth and both documents and tests what it observes.
- **No behavior change.** This is documentation plus a guard test. The runtime
  invocation order, the set of populated fields, and `Singleton` handling are
  described as-is, not altered. If the implementing agent finds a genuine
  liveness bug while writing the guard test (e.g. a field the contract should
  populate but doesn't), that is surfaced as a separate finding, not silently
  "fixed" under this PRD.
- **Single source of truth.** The seam (the struct doc) is authoritative; the
  developer docs reference/mirror it rather than restating a second, driftable
  copy. The guard test pins the seam to `Base`.
- **`CrudCtx` field semantics are clarified in place.** The existing inline
  comments (`Plan` = typed plan, null otherwise; `State` = typed prior state;
  `Body` = mutate in `Before*`; `Response` = read in `After*`) are made precise
  about *which* hook points make each "otherwise null" — i.e. the "otherwise" is
  enumerated, not left to the reader.

## Testing Decisions

- **What makes a good test here.** It asserts the *external contract a hook author
  relies on* — "when hook point X fires, fields A and B are live and C and D are
  null, and X runs before the GET/PUT" — by driving the real `Base` CRUD methods
  and observing the `CrudCtx` each hook receives. It does **not** assert internal
  call sequencing of unexported helpers, and it does **not** merely check that a
  hook is registered (the existing `*HookRegistered` tests already cover
  registration and are not what this guards).
- **Seam (approved).** The highest available seam is `Base`'s CRUD entry points
  (`Create`/`Read`/`Update`/`Delete`) invoked in-process with a recording hook
  that captures the `CrudCtx` it was handed. Assertions are made on that captured
  context's field liveness. This is one seam, already proven usable.
- **Prior art to follow.** `internal/native/resource/base_read_test.go`
  (`TestReadFiresBeforeReadHookBeforeGet`, added with the `BeforeRead` wiring)
  is the template: it builds a valid framework state, drives `Base.Read`
  end-to-end, and asserts the hook ran and observed the decoded state — plus a
  recording transport proving ordering relative to the GET. The harness in
  `base_import_test.go` (`fakeCred`, `fakeTransport`/`recordingTransport`,
  `fakeClient`/`recordingClient`, `stringObject`, `hasErrorSummary`, and the
  `resourceGroupReadState` state-builder) is reused and extended to the other
  hook points. The resource-group descriptor is the simplest driver; storage
  account/blob service are available where a `Singleton` path must be exercised.
- **Modules tested.** `internal/native/resource` (`Base` + `Hooks`/`CrudCtx`).
- **Coverage target.** One assertion group per hook point `Base` invokes, plus the
  `Singleton` suppression rule (a `Singleton` delete does not call
  `BeforeDelete`). Test authoring is delegated to the Tester agent per repo
  convention.

## Out of Scope

- Wiring or removing `BeforeRead` — already done and committed.
- Any change to hook *invocation order*, the set of populated `CrudCtx` fields, or
  `Singleton` semantics. This PRD documents and guards current behavior; it does
  not redesign it.
- Adding new hook points, splitting `Hooks` into create/read/delete-specific
  context types, or introducing per-hook-point typed contexts. That is a larger
  interface redesign; if the field-liveness matrix turns out to be the real pain,
  note it for a follow-up, but do not build it here.
- Live-Azure acceptance tests. The contract is verifiable in-process; the Azure
  suite is unaffected.

## Further Notes

- This is the smallest-blast-radius of the three "worth exploring" candidates and
  the natural continuation of the `BeforeRead` fix already shipped: the same seam,
  the same test harness, finishing the job of making the hook contract honest.
- The guard test is the load-bearing part. Documentation alone drifts (that is how
  the `BeforeRead` contradiction arose); the test is what keeps the seam and
  `Base` in agreement going forward.
- No ADR is reopened. This preserves ADR-0005 (azapin knowledge is compiled in)
  and the schema/runtime split of ADR-0002/ADR-0007 — hooks customize runtime
  behavior only, never the schema.
