// Package validators holds the hand-written schema validators that are generic
// and cross-resource — semantic rules a bicep type or the azwise overlay cannot
// express but that are not tied to any one service (a UUID, an ARM resource ID).
// One validator per file, each an exported constructor returning a
// validator.String. Resource-specific validators do not live here; they belong in
// internal/native/services/<service>/validators.
//
// A customizer references the constructor by its real Go symbol —
// typegraph.Validator(validators.UUID) — so a rename or a deletion is a compile
// error at the callsite, not a silent bad string. The generator never imports
// this package: the emitter reflects the passed func value at generation time
// (runtime.FuncForPC) to recover both the qualified call (validators.UUID()) and
// this package's import path, and bakes them into the generated schema, which is
// what actually links against the constructor.
package validators
