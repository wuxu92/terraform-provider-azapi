// Package planmodifiers holds the hand-written schema plan modifiers that are
// generic and cross-resource — drift-suppression and plan-shaping rules the
// built-in framework vocabulary cannot express but that are not tied to any one
// service (equivalent-location / equivalent-resource-ID reuse, the discriminated
// -variant modifier). One modifier per file, each an exported constructor
// returning the framework planmodifier.String / .Object it applies to.
// Resource-specific plan modifiers do not live here; they belong in
// internal/native/services/<service>/planmodifiers.
//
// A customizer references the constructor by its real Go symbol —
// typegraph.PlanModifier(planmodifiers.UseStateForEquivalentResourceID) — so a
// rename or a deletion is a compile error at the callsite, not a silent bad
// string. The generator never imports this package: the emitter reflects the
// passed func value at generation time (runtime.FuncForPC) to recover both the
// qualified call (planmodifiers.UseStateForEquivalentResourceID()) and this
// package's import path, and bakes them into the generated schema, which is what
// actually links against the constructor. The built-in location and
// discriminated-variant modifiers the emitter emits unconditionally are qualified
// the same way.
package planmodifiers
