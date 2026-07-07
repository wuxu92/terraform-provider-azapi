# PRD: Auto-register acceptance scope teardown at the seam

Status: ready-for-agent

_Source: architecture review of `internal/native` (2026-07-06), candidate 7 — "Pull the Workspace scope-teardown invariant into the seam"._

## Problem Statement

The azapin acceptance harness (`nativeacc`) is a genuinely deep test surface:
`Workspace` concentrates provider-server reattach, apply/destroy orchestration,
drift and existence checks, and per-scope resource ownership behind a small
scenario interface. Authors declare a base of shared resources in a root
**scope** and carve nested scopes (`ws.Scope()` / `Scope.Scope()`) for the
resources of nested Ginkgo `Ordered` containers, so a nested container reuses the
base its enclosing containers built, referencing them by Terraform address.

One contract, though, leaks out of that seam and onto the author's discipline:

- **Teardown must be hand-wired.** Every `ws.Scope()` must be paired, by hand,
  with `AfterAll(scope.Teardown)` in the same container. This pairing is pure
  boilerplate — the review found it duplicated at every scope site — and it is
  invisible to the seam: nothing about `ws.Scope()` requires or performs it.
- **A forgotten pairing silently leaks Azure resources.** If an author creates a
  scope and forgets its `AfterAll(scope.Teardown)`, that scope's resources are
  applied to Azure and never destroyed. There is no compile error, no test
  failure — just real cloud resources left running and billing.
- **Destroy order rides on nesting discipline.** Dependents must be destroyed
  before the resources they depend on. Today that ordering is achieved only by the
  author correctly nesting a dependent's scope *deeper* than its dependency's, so
  Ginkgo's inner-`AfterAll`-before-outer semantics happen to destroy inner first.
  The ordering invariant is real but is enforced by author care, not by the seam.

So the deep, well-factored part of the harness is undermined by one
discipline-dependent step whose only failure signal is a leaked cloud resource
discovered later.

## Solution

Move the teardown contract **into the seam** so the harness, not the author's
memory, guarantees it:

1. **Creating a scope auto-registers its teardown.** `ws.Scope()` /
   `Scope.Scope()` register `Teardown` with the enclosing `Ordered` container
   itself, so the author who asks for a child scope gets its destroy wired
   automatically. The routine `AfterAll(scope.Teardown)` line disappears from
   consumers.
2. **Destroy order is derived from the scope tree, not restated.** Because each
   child scope registers its own teardown at its own container, the existing
   inner-container-first execution destroys dependents before dependencies by
   construction of the scope nesting — the author expresses the dependency by
   nesting (which they already do to reference ancestors), and ordering falls out
   of that single fact rather than a second, separately-maintained wiring step.
3. **A misuse fails loudly, not silently.** Asking for a child scope outside an
   `Ordered` container (where teardown can't be registered) is a loud, immediate
   error, not a silent leak — and a safety-net assertion can confirm every scope
   the workspace handed out had its teardown registered.

The author writes less, forgets nothing, and can no longer leak a scope's
resources by omission.

## User Stories

1. As an acceptance-test author, I want asking for a child scope to wire its
   teardown automatically, so that I can't forget `AfterAll(scope.Teardown)` and
   leak Azure resources.
2. As an acceptance-test author, I want to stop writing the boilerplate
   `AfterAll(scope.Teardown)` line at every scope, so that scope declaration is one
   call, not two coupled calls.
3. As an acceptance-test author, I want a dependent's scope nested inside its
   dependency's scope to destroy the dependent first, so that teardown respects
   the dependency graph without me wiring order explicitly.
4. As an acceptance-test author, I want creating a scope in an unsupported place
   (outside an `Ordered` container) to fail immediately and loudly, so that the
   mistake surfaces at development time rather than as a leaked resource.
5. As a maintainer, I want the teardown invariant enforced by the harness seam, so
   that a new test author inherits correct teardown without reading the workspace
   internals.
6. As a maintainer, I want the destroy ordering derived from the scope tree, so
   that there is one source of truth (the nesting) instead of a separate,
   drift-prone ordering step.
7. As a maintainer, I want a fast in-process test proving that creating a scope
   registers its teardown, so that the contract is guarded without a live-Azure
   run.
8. As a maintainer, I want a test proving nested scopes register teardown in
   inner-first order, so that the dependency-safe destroy sequence is pinned.
9. As a maintainer, I want the auto-registered teardown to remain a no-op on a
   skipped run (workspace never started), so that unit and skipped acceptance runs
   don't attempt Azure calls.
10. As a maintainer, I want the root scope's resources to keep being torn down by
    the workspace's whole-workspace destroy, so that only the child-scope
    boilerplate is removed and the base lifecycle is unchanged.
11. As a maintainer, I want the change to preserve every existing acceptance
    scenario's behavior (same resources applied, same destroy outcomes), so that
    the harness improvement doesn't perturb live results.
12. As an acceptance-test author, I want the existing scoped-teardown semantics
    (destroy only this scope's resources by removing their files and re-applying;
    assert they're gone from Azure) preserved, so that auto-registration changes
    *who wires* teardown, not *what teardown does*.
13. As a maintainer, I want a safety-net check that every scope the workspace
    vended had its teardown registered, so that any path that bypasses
    auto-registration is caught rather than leaking.
14. As a maintainer, I want the migration to update the existing acceptance suites
    to the auto-registering form, so that no consumer is left double-wiring
    teardown.

## Implementation Decisions

- **Module touched.** `internal/native/acceptance` (`nativeacc`) — `Workspace`
  and `Scope` (scope creation + teardown registration), plus its in-process unit
  tests. The consuming acceptance suites in the generated service packages
  (`resources`, `storage`, `web` `*_test.go`) are migrated to drop their now-
  redundant `AfterAll(scope.Teardown)` lines.
- **Auto-registration uses Ginkgo, already a dependency.** `nativeacc` already
  imports `github.com/onsi/ginkgo/v2`; `ws.Scope()` / `Scope.Scope()` register the
  child scope's `Teardown` via Ginkgo's container-level cleanup (an `AfterAll`- or
  `DeferCleanup`-equivalent node) at construction time, which is when scopes are
  created (in the `Describe`/`Context` body). No new dependency is introduced.
- **Registration goes through an injectable seam.** The call that registers
  teardown is routed through a single indirection (a package-level function
  variable / a `Workspace` field defaulting to the Ginkgo registrar) so that an
  in-process test can substitute a recorder and observe registration without
  running Ginkgo or Terraform. This is the one seam the feature is tested at.
- **Ordering is a property of the scope tree, not a stored field.** No explicit
  order list is maintained. Each child scope registers its own teardown at its own
  container; Ginkgo's existing inner-before-outer `AfterAll` execution yields
  dependent-first destroy because dependents are nested deeper (the same nesting
  authors already use to reference ancestor resources). The PRD does not add a new
  ordering mechanism; it removes the need to hand-wire the one that nesting already
  implies.
- **Teardown behavior is unchanged.** `Scope.Teardown` keeps its current
  semantics: remove this scope's resource files, re-apply so Terraform destroys
  exactly those, assert each tracked resource is absent from Azure, and no-op when
  the workspace never started. Only the *registration* of that method moves from
  the consumer into `ws.Scope()`.
- **Root scope unchanged.** Root-scope base resources continue to be destroyed by
  the workspace-wide `Destroy` (wired at the top container's `AfterAll`).
  Auto-registration applies to child scopes from `ws.Scope()` / `Scope.Scope()`.
- **Loud failure over silent leak.** If a child scope is requested where teardown
  cannot be registered (outside an `Ordered` container), the harness fails
  immediately rather than proceeding to apply un-torn-down resources. Optionally,
  a workspace-level safety-net assertion confirms every vended scope had its
  teardown registered.
- **No API surface removed.** `Scope.Teardown` stays exported (the safety net and
  any bespoke ordering still call it); the change makes its wiring automatic, it
  does not delete the method.

## Testing Decisions

- **What makes a good test here.** It asserts the **observable contract**:
  "creating a child scope registers exactly one teardown for that scope" and
  "nested child scopes register teardown in inner-first order," observed through
  the injected registrar seam — not the internals of Ginkgo or Terraform. It also
  keeps the existing `Scope` ownership/dedup guarantees green. It must **not**
  require Azure or a Terraform binary; the registration contract is provable
  in-process.
- **Seam (approved).** The single injectable teardown-registrar indirection is the
  seam. A unit test swaps in a recorder and drives `ws.Scope()` / `Scope.Scope()`
  directly, asserting what got registered and in what nesting order. This is the
  highest, narrowest seam — one indirection all scope creation flows through — and
  is introduced at the point of registration.
- **Prior art to follow.** `internal/native/acceptance/scope_test.go` already
  exercises `Scope` in-process without Azure (`TestScopeOwnKeyedByAddress`,
  `TestScopeOwnRejectsLabelCollision`, `TestResourceFileNameKeyedByAddress`,
  `TestStageBundlesConfigAndChecks`). The new tests live beside them and follow the
  same white-box, no-network pattern, using the recorder seam in place of the live
  Ginkgo registrar.
- **Modules tested.** `internal/native/acceptance` (`Workspace`/`Scope` scope
  creation + teardown registration).
- **Coverage target.** (a) `ws.Scope()` registers one teardown for the created
  scope; (b) `Scope.Scope()` on a child registers at the deeper level so execution
  order is inner-first; (c) requesting a scope where registration is impossible
  fails loudly; (d) the safety-net assertion flags a vended scope whose teardown
  was never registered; (e) existing `Scope` ownership tests still pass. Test
  authoring is delegated to the Tester agent per repo convention. The live-Azure
  acceptance suites remain the end-to-end proof that resources are actually
  destroyed.

## Out of Scope

- Any change to **what** `Scope.Teardown` does (file removal + re-apply + Azure
  absence assertion) or to the workspace-wide `Destroy`. This PRD moves *where
  teardown is registered*, not its behavior.
- Redesigning `Workspace`/`Scope` more broadly. The review is explicit that
  `Workspace` is a deep module to preserve; this narrows exactly one leaking
  contract and touches nothing else.
- Changing how resources are declared, applied, staged (`ApplyAll`/`Stage`), or
  checked, or how ancestor references by Terraform address work.
- Auto-deriving scope nesting from resource dependencies (i.e. inferring the tree
  from `IDRef` usage). Nesting stays author-expressed; only teardown wiring is
  automated. Inferring the tree is a larger idea to note, not build here.
- The provider-server reattach, drift, and existence-check machinery — unaffected.

## Further Notes

- This is the highest-leverage of the three "worth exploring" candidates for
  *correctness of the test surface*: the failure mode it removes is leaked, billed
  Azure resources from a forgotten one-liner — a silent, expensive class of bug
  that discipline alone will eventually miss.
- The change is deliberately small: one indirection, auto-registration in
  `ws.Scope()`/`Scope.Scope()`, and a consumer migration that only *deletes*
  boilerplate. Preserving `Teardown`'s behavior byte-for-byte keeps the live
  acceptance results stable.
- No ADR is reopened. This is confined to the acceptance harness and does not
  touch the generation pipeline (ADR-0007), the schema/runtime split
  (ADR-0002/ADR-0005), or verification (ADR-0008).
