package nativeacc

import "testing"

// TestResourceFileNameKeyedByAddress guards against resource .tf files colliding when
// two resource types share a Terraform label: the file name is keyed by the full
// tfType.label address, not the label alone, so it stays unique across types.
func TestResourceFileNameKeyedByAddress(t *testing.T) {
	sa := resourceFileName("azapi_storage_account.test")
	rg := resourceFileName("azapi_resource_group.test")
	if sa == rg {
		t.Fatalf("same-label/different-type resources share a file name: %q", sa)
	}
	if want := "resource_azapi_storage_account.test.tf"; sa != want {
		t.Errorf("file name = %q, want %q", sa, want)
	}
}

// TestScopeOwnKeyedByAddress guards that a scope tracks owned files by the owning
// resource's address: distinct addresses are kept (even with a shared label across
// types, the case that motivated the file-keying fix), while re-applying the SAME
// resource records it once so Teardown removes each file exactly once.
func TestScopeOwnKeyedByAddress(t *testing.T) {
	var s Scope
	sa := &Resource{scope: &s, tfType: "azapi_storage_account", label: "test"}
	rg := &Resource{scope: &s, tfType: "azapi_resource_group", label: "test"}
	s.own(sa)
	s.own(rg) // same label, different type — must stay distinct
	s.own(sa) // same resource re-applied — deduped
	if len(s.owned) != 2 {
		t.Fatalf("own should retain 2 distinct addresses, got %d: %v", len(s.owned), s.owned)
	}
}

// TestScopeOwnRejectsLabelCollision guards Option C's safety net: two DISTINCT
// resources that land on the same tfType.label address in one scope are a label
// collision (a second instance built without a distinct label) and must panic loudly
// rather than silently clobber the first resource's .tf file.
func TestScopeOwnRejectsLabelCollision(t *testing.T) {
	var s Scope
	first := &Resource{scope: &s, tfType: "azapi_storage_account", label: "test"}
	dup := &Resource{scope: &s, tfType: "azapi_storage_account", label: "test"}
	s.own(first)
	defer func() {
		if recover() == nil {
			t.Fatal("own should panic on a duplicate address from a distinct resource")
		}
	}()
	s.own(dup)
}

// TestStringConfigure adapts raw HCL into the Configure interface without keeping a
// separate config resolver in Resource.Apply.
func TestStringConfigure(t *testing.T) {
	const hcl = "resource azapi_resource_group rg {}"
	var cfg Configure = StringConfigure(hcl)
	if got := cfg.Config(); got != hcl {
		t.Errorf("StringConfigure.Config() = %q, want %q", got, hcl)
	}
}

// TestStageBundlesConfigAndChecks guards Resource.Stage, the batched-apply building
// block: it must bundle the handle, config and checks verbatim so ApplyAll writes the
// right config and runs the right checks per resource. The apply itself drives
// Terraform and is exercised by the Azure-gated acceptance suite.
func TestStageBundlesConfigAndChecks(t *testing.T) {
	var s Scope
	r := &Resource{scope: &s, tfType: "azapi_resource_group", label: "rg"}
	chk := func(*checkCtx) {}
	cfg := StringConfigure("cfg-hcl")
	st := r.Stage(cfg, chk, chk)
	if st.r != r {
		t.Errorf("Stage dropped the resource handle")
	}
	if st.config != cfg {
		t.Errorf("Stage config = %v, want %q", st.config, cfg)
	}
	if len(st.checks) != 2 {
		t.Errorf("Stage checks = %d, want 2", len(st.checks))
	}
}

// TestScopeCreationRegistersTeardown proves the core auto-registration contract:
// asking the workspace for a child scope routes that exact scope through the teardown
// registrar once, tracks it in vended, and marks it registered — so the author gets
// the child's destroy wired without a paired AfterAll(child.Teardown) line.
func TestScopeCreationRegistersTeardown(t *testing.T) {
	ws := NewWorkspace()
	var recorded []*Scope
	ws.registerTeardown = func(child *Scope) {
		recorded = append(recorded, child)
		child.teardownRegistered = true
	}

	child := ws.Scope()

	if len(recorded) != 1 {
		t.Fatalf("registrar invoked %d times, want exactly 1", len(recorded))
	}
	if recorded[0] != child {
		t.Errorf("registered scope %p is not the scope returned to the caller %p", recorded[0], child)
	}
	if len(ws.vended) != 1 || ws.vended[0] != child {
		t.Errorf("vended = %v, want exactly the returned child %p", ws.vended, child)
	}
	if !child.teardownRegistered {
		t.Error("child.teardownRegistered = false, want true once the registrar wired it")
	}
}

// TestNestedScopeRegistersInnerFirstOrder proves Scope.Scope() on a child registers the
// deeper scope through the same seam, in outer-then-inner construction order.
//
// Why this pins inner-first destroy: each scope registers its Teardown at its OWN
// Ordered container as that container is constructed, and Ginkgo runs a deeper
// container's AfterAll before its enclosing one. So the inner scope registered here —
// at the deeper container — is torn down BEFORE the outer it nests inside, which is
// exactly the dependent-first order authors get for free by nesting. We cannot (and
// must not) run Ginkgo in a unit test, so we observe the registration/nesting order
// through the seam, which is precisely what determines that teardown order.
func TestNestedScopeRegistersInnerFirstOrder(t *testing.T) {
	ws := NewWorkspace()
	var recorded []*Scope
	ws.registerTeardown = func(child *Scope) {
		recorded = append(recorded, child)
		child.teardownRegistered = true
	}

	outer := ws.Scope()
	inner := outer.Scope()

	if len(recorded) != 2 {
		t.Fatalf("registrar invoked %d times, want 2 (outer + inner)", len(recorded))
	}
	if recorded[0] != outer {
		t.Errorf("first registration = %p, want outer %p", recorded[0], outer)
	}
	if recorded[1] != inner {
		t.Errorf("second registration = %p, want inner %p", recorded[1], inner)
	}
	if inner.ws != outer.ws {
		t.Error("inner.ws != outer.ws: a nested scope must share the workspace so it can reference ancestor resources by address")
	}
}

// TestScopeCreationFailsLoudlyWhenRegistrationImpossible proves scope creation does not
// swallow a registrar failure. In production, ws.Scope() outside an Ordered container
// makes ginkgo.AfterAll abort; here we model that impossible registration with a
// registrar that panics. Scope() must propagate the panic loudly rather than hand back
// a scope whose teardown silently never wired — which would leak Azure resources.
func TestScopeCreationFailsLoudlyWhenRegistrationImpossible(t *testing.T) {
	ws := NewWorkspace()
	ws.registerTeardown = func(*Scope) { panic("no Ordered container") }
	defer func() {
		if recover() == nil {
			t.Fatal("Scope() should propagate the registrar panic, not swallow it and return an unwired scope")
		}
	}()
	ws.Scope()
}

// TestSafetyNetFlagsUnregisteredScope proves the whole-workspace safety net
// (assertScopesRegistered, called at the top of Destroy) fires exactly on a real leak:
// a scope that was vended but never had its Teardown registered. The negative case
// records without marking, so the net must panic; the positive case marks, so the same
// net is a no-op. Asserting both proves the net catches the leak without crying wolf on
// a correctly-registered workspace.
func TestSafetyNetFlagsUnregisteredScope(t *testing.T) {
	// Negative: recorder vends the scope but does NOT set teardownRegistered, so
	// ws.vended holds an unregistered scope — the leak the net exists to catch.
	ws := NewWorkspace()
	var recorded []*Scope
	ws.registerTeardown = func(child *Scope) { recorded = append(recorded, child) }
	ws.Scope()
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("assertScopesRegistered should panic on a vended-but-unregistered scope")
			}
		}()
		ws.assertScopesRegistered()
	}()

	// Positive: a marking recorder leaves the vended scope registered, so the same
	// safety net is a no-op. If it panics here, the test fails (unrecovered panic).
	okWS := NewWorkspace()
	okWS.registerTeardown = func(child *Scope) { child.teardownRegistered = true }
	okWS.Scope()
	okWS.assertScopesRegistered()
}
