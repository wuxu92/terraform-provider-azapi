package typegraph

// This file is the developer-facing fluent vocabulary for per-resource
// customizers. A customizer runs at generation time (see the
// generator/customizers sub-package) and mutates the resolved type graph and the
// operational-envelope spec in place before emission. Doing that by hand means
// the same three-line dance over and over — FindProperty(def, path), then flip a
// flag or append a validator — repeated for every path and every resource.
//
// The methods below collapse that dance into intent-named, path-variadic calls
// on *ResourceDefinition. Each targets body properties by ARM dot path (the same
// paths FindProperty understands) or the operational envelope, and each returns
// the definition so calls chain. Like FindProperty, every path-based method
// PANICS on an unresolved path: a customizer authors paths by hand, so a typo
// must fail generation immediately rather than be silently skipped.
//
// Array-element note: sibling arrays that share a deduplicated bicep element type
// (e.g. ipRules and ipv6Rules) still require an IsolateArrayElement call before
// flagging one array's elements, otherwise the change leaks to the siblings.
// These helpers operate on whatever the path resolves to, so isolate first, then
// address the element property by its full path.

// AddValidatorsFor appends validators to the body property at armPath. It is the
// method form of `FindProperty(def, armPath).Validators = append(...)`, for
// attaching semantic rules a bicep type cannot express (resource-ID shape, a
// resource-specific regex, an element enum). Panics if armPath does not resolve.
func (def *ResourceDefinition) AddValidatorsFor(armPath string, validators ...DescriptionValidator) *ResourceDefinition {
	if len(validators) == 0 {
		return def
	}
	p := FindProperty(def, armPath)
	p.Validators = append(p.Validators, validators...)
	return def
}

// Required promotes every body property at the given ARM paths to Required
// (FlagRequired). Use it when AzureRM requires a field the bicep graph leaves
// optional (azwise RequiredFields is planner metadata, not a schema flag), so a
// malformed body fails at plan time instead of as an ARM 4xx. Panics on any path
// that does not resolve.
func (def *ResourceDefinition) Required(paths ...string) *ResourceDefinition {
	return def.forEachProperty(paths, func(p *Property) { p.Flags |= FlagRequired })
}

// Computed forces every body property at the given ARM paths to Computed-only
// (ForceComputed), the equivalent of an azwise ComputedFields entry. Use it for a
// field the server always owns that the bicep graph leaves settable. Panics on
// any path that does not resolve.
func (def *ResourceDefinition) Computed(paths ...string) *ResourceDefinition {
	return def.forEachProperty(paths, func(p *Property) { p.ForceComputed = true })
}

// ForceNew marks every body property at the given ARM paths for replacement
// (emits a RequiresReplace plan modifier), the equivalent of an azwise ForceNew
// entry. Use it for an immutable field the overlay does not already cover. Panics
// on any path that does not resolve.
func (def *ResourceDefinition) ForceNew(paths ...string) *ResourceDefinition {
	return def.forEachProperty(paths, func(p *Property) { p.ForceNew = true })
}

// Sensitive marks every body property at the given ARM paths sensitive (emits
// Sensitive: true), the equivalent of an azwise SensitiveFields entry. Panics on
// any path that does not resolve.
func (def *ResourceDefinition) Sensitive(paths ...string) *ResourceDefinition {
	return def.forEachProperty(paths, func(p *Property) { p.Sensitive = true })
}

// AsSet emits every primitive-array property at the given ARM paths as a
// SetAttribute instead of a ListAttribute (UseSet). Use it only when ARM treats
// the collection as unordered (e.g. CORS header names) so API reordering does not
// produce drift. Panics on any path that does not resolve.
func (def *ResourceDefinition) AsSet(paths ...string) *ResourceDefinition {
	return def.forEachProperty(paths, func(p *Property) { p.UseSet = true })
}

// WithEmptyListDefault sets an empty-list schema default (DefaultEmptyList) on
// every Optional+Computed array property at the given ARM paths. Use it when ARM
// always echoes an omitted list back as [] rather than absent: the default makes
// the omitted-config plan value ([]) match the API on apply/read/import, while an
// explicitly-configured list is left untouched. Panics on any path that does not
// resolve.
func (def *ResourceDefinition) WithEmptyListDefault(paths ...string) *ResourceDefinition {
	return def.forEachProperty(paths, func(p *Property) { p.DefaultEmptyList = true })
}

// NonNullStateForUnknown opts every computed property at the given ARM paths out
// of the default UseStateForUnknown in favour of UseNonNullStateForUnknown: a
// null prior plans as "(known after apply)" so the server may populate it. Use it
// for server-controlled read-only fields that transition null -> non-null when a
// sibling/parent is configured. Panics on any path that does not resolve.
func (def *ResourceDefinition) NonNullStateForUnknown(paths ...string) *ResourceDefinition {
	return def.forEachProperty(paths, func(p *Property) { p.NonNullStateForUnknown = true })
}

// Default sets the schema default of the body property at armPath, overriding any
// description-mined default. Panics if armPath does not resolve.
func (def *ResourceDefinition) Default(armPath, value string) *ResourceDefinition {
	FindProperty(def, armPath).DefaultValue = value
	return def
}

// SetNameValidators replaces the operational-envelope name attribute's validators
// with the given set. The ARM resource name is not part of the bicep body graph
// (it maps to the ARM ID), so a resource's naming constraint is attached here
// rather than via the azwise overlay, which skips empty-path (name) rules.
func (def *ResourceDefinition) SetNameValidators(validators ...DescriptionValidator) *ResourceDefinition {
	def.Envelope.Name.Validators = validators
	return def
}

// SetParent overrides the operational-envelope parent-reference attribute name
// and description (e.g. renaming the generic parent_id to scope_id for an
// extension resource). The parent ID-shape validator seeded by PostProcess is
// left intact.
func (def *ResourceDefinition) SetParent(name, description string) *ResourceDefinition {
	def.Envelope.Parent.Name = name
	def.Envelope.Parent.Description = description
	return def
}

// AddMetaAttr appends synthetic, behavior-only top-level attributes to the
// operational envelope. A MetaAttr is not part of the bicep body graph and never
// travels to ARM; it drives provider-side behavior a runtime hook reads from
// state (e.g. purge_on_destroy). Emitted as an Optional bool schema attribute.
func (def *ResourceDefinition) AddMetaAttr(attrs ...MetaAttr) *ResourceDefinition {
	def.Envelope.Meta = append(def.Envelope.Meta, attrs...)
	return def
}

// forEachProperty resolves each ARM path to its body property and applies fn.
// FindProperty panics on an unresolved path, so a customizer typo fails
// generation rather than being silently skipped.
func (def *ResourceDefinition) forEachProperty(paths []string, fn func(*Property)) *ResourceDefinition {
	for _, path := range paths {
		fn(FindProperty(def, path))
	}
	return def
}
