// Package validators holds the hand-written schema validators specific to
// Microsoft.Network native resources — semantic rules a bicep type or the azwise
// overlay cannot express and that are not generic enough to live in the shared
// internal/native/schema/validators package. One validator per file, each an
// exported constructor returning a validator.String.
//
// A customizer references the constructor by its real Go symbol —
// typegraph.Validator(networkvalidators.VirtualNetworkBgpCommunity) — so a rename
// or a deletion is a compile error at the callsite, not a silent bad string. The
// generator never imports this package: the emitter reflects the passed func value
// at generation time (runtime.FuncForPC) to recover both the qualified call and
// this package's import path, and bakes them into the generated schema.
package validators
