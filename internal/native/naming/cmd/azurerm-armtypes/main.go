// Command azurerm-armtypes mechanically extracts the mapping from ARM resource
// type (e.g. "Microsoft.DocumentDB/databaseAccounts") to the AzureRM resource
// noun (e.g. "azurerm_cosmosdb_account") by tracing, for every resource the
// terraform-provider-azurerm registers, the ID value its Create path builds:
//
//	azurerm resource  ->  Create func  ->  id := <alias>.New<Thing>ID(...)
//	                  ->  <alias> resolved via imports to a go-azure-sdk /
//	                      go-azure-helpers commonids package
//	                  ->  that package's fmtString "/subscriptions/%s/.../providers/
//	                      Microsoft.X/types/%s/..."  ->  ARM type.
//
// AzureRM's resource nouns are the human-curated authority for how practitioners
// already know each Azure resource, so the azapi native generator adopts them as
// its naming reference (see internal/native/naming). This tool regenerates that
// reference table. The table is a naming HINT for the generator: an imperfect
// entry can be patched in the naming package, so the heuristics below favour
// resolving obvious families over being conservative.
//
// When several AzureRM resources share one ARM type, the group is auto-resolved
// where the members have an obvious shared identity:
//
//   - main resource: one member is a whole-word prefix of every other
//     (Microsoft.ApiManagement/service -> azurerm_api_management is the parent of
//     azurerm_api_management_policy, ...); the parent wins.
//   - discriminated base: every member is a common word-prefix plus a distinct
//     suffix (.../identityProviders -> azurerm_api_management_identity_provider_
//     {aad,aadb2c,facebook,...}; .../credentials -> azurerm_data_factory_
//     credential_{service_principal,user_managed_identity}); the shared prefix is
//     used, unless it already names another resource.
//
// Groups fitting neither pattern (Microsoft.Web/sites -> linux/windows web_app
// and function_app) stay ambiguous, and resources whose Create uses a
// legacy/internal ID parser are unresolved; both are listed in the report and
// fall back to the mechanical naming rule rather than being guessed. Run:
//
//	go run ./internal/native/naming/cmd/azurerm-armtypes \
//	  -azurerm ../terraform-provider-azurerm \
//	  -out internal/native/naming/azurerm_reference_gen.go \
//	  -report internal/native/naming/azurerm_reference_report.md
package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func main() {
	azurermPath := flag.String("azurerm", "../terraform-provider-azurerm", "path to terraform-provider-azurerm repo root")
	outPath := flag.String("out", "internal/native/naming/azurerm_reference_gen.go", "generated Go file path")
	reportPath := flag.String("report", "internal/native/naming/azurerm_reference_report.md", "skip/ambiguity report path")
	flag.Parse()

	sdkIndex, err := buildSDKIndex(*azurermPath)
	if err != nil {
		fatalf("build SDK index: %v", err)
	}
	fmt.Fprintf(os.Stderr, "SDK id index: %d constructors across %d packages\n", sdkIndex.count(), sdkIndex.pkgCount())

	resources, skips := extractResources(*azurermPath, sdkIndex)

	// Invert name->armType into armType->names to detect ambiguity. ARM types are
	// case-insensitive, so group by a lowercased key: two AzureRM resources whose
	// IDs spell the same type with different casing (e.g. serverFarms vs
	// serverfarms) are the same ARM type and must not both survive as "unique".
	// The first-seen original casing is kept for the emitted (readable) key.
	byARM := map[string][]string{}
	origCasing := map[string]string{}
	for name, armType := range resources {
		key := strings.ToLower(armType)
		byARM[key] = append(byARM[key], name)
		if _, ok := origCasing[key]; !ok {
			origCasing[key] = armType
		}
	}

	type entry struct {
		armType string
		name    string
	}
	type resolvedGroup struct {
		armType string
		name    string
		members []string
		kind    string
	}

	// First pass: unique ARM types own their name outright. Collect those names
	// into taken so a later discriminated-base collapse cannot duplicate a name a
	// real resource already holds (e.g. .../dataflows must not collapse to
	// azurerm_data_factory, which already names Microsoft.DataFactory/factories).
	var clean []entry
	taken := map[string]bool{}
	var ambiguousKeys []string
	for key, names := range byARM {
		if len(names) == 1 {
			clean = append(clean, entry{origCasing[key], names[0]})
			taken[names[0]] = true
			continue
		}
		ambiguousKeys = append(ambiguousKeys, key)
	}

	// Second pass: compute a candidate resolution for each ambiguous group, then
	// accept them in confidence order — a main resource is a stronger signal than a
	// discriminated base — rejecting any base a unique resource or an
	// already-accepted group claims. Two distinct ARM types must never resolve to
	// the same azapi noun, so a discriminated base contested by more than one group
	// is dropped for all of them; the losers fall back to the mechanical rule (the
	// four RecoveryServices replication* types that share only azurerm_site_recovery,
	// and Microsoft.Insights/webTests, which would otherwise shadow the
	// azurerm_application_insights that .../components legitimately owns as a main).
	type candidate struct {
		key, base, kind string
		names           []string
	}
	var autoResolved []resolvedGroup
	var ambiguous []string
	var mains, discs []candidate
	for _, key := range ambiguousKeys {
		names := byARM[key]
		sort.Strings(names)
		base, kind, ok := resolveAmbiguous(names, taken)
		if !ok {
			ambiguous = append(ambiguous, fmt.Sprintf("%s -> %s", origCasing[key], strings.Join(names, ", ")))
			continue
		}
		c := candidate{key, base, kind, names}
		if kind == "main resource" {
			mains = append(mains, c)
		} else {
			discs = append(discs, c)
		}
	}
	sort.Slice(mains, func(i, j int) bool { return mains[i].key < mains[j].key })
	sort.Slice(discs, func(i, j int) bool { return discs[i].key < discs[j].key })

	// A discriminated base wanted by more than one group is inherently ambiguous.
	discWant := map[string]int{}
	for _, c := range discs {
		discWant[c.base]++
	}
	accept := func(c candidate) {
		taken[c.base] = true
		autoResolved = append(autoResolved, resolvedGroup{origCasing[c.key], c.base, c.names, c.kind})
		clean = append(clean, entry{origCasing[c.key], c.base})
	}
	reject := func(c candidate) {
		ambiguous = append(ambiguous, fmt.Sprintf("%s -> %s", origCasing[c.key], strings.Join(c.names, ", ")))
	}
	for _, c := range mains {
		if taken[c.base] {
			reject(c)
			continue
		}
		accept(c)
	}
	for _, c := range discs {
		if taken[c.base] || discWant[c.base] > 1 {
			reject(c)
			continue
		}
		accept(c)
	}
	sort.Slice(clean, func(i, j int) bool { return clean[i].armType < clean[j].armType })
	sort.Slice(autoResolved, func(i, j int) bool { return autoResolved[i].armType < autoResolved[j].armType })
	sort.Strings(ambiguous)

	// Emit generated table.
	var b strings.Builder
	b.WriteString("// Code generated by internal/native/naming/cmd/azurerm-armtypes; DO NOT EDIT.\n")
	b.WriteString("//\n")
	b.WriteString("// Source of truth: terraform-provider-azurerm resource registrations, traced\n")
	b.WriteString("// through each resource's Create ID constructor to the go-azure-sdk fmtString.\n")
	b.WriteString("// ARM types with a unique AzureRM resource, plus ambiguous groups auto-resolved\n")
	b.WriteString("// to a main resource or a discriminated base, are listed here; still-ambiguous\n")
	b.WriteString("// and unresolved types are omitted (see azurerm_reference_report.md) and fall\n")
	b.WriteString("// back to the mechanical naming rule in naming.go.\n\n")
	b.WriteString("package naming\n\n")
	b.WriteString("// azurermResourceForARMType maps a bare ARM resource type to the AzureRM\n")
	b.WriteString("// resource noun that authoritatively names it. ResourceName consults this\n")
	b.WriteString("// first (case-insensitively), swapping the azurerm_ prefix for azapi_.\n")
	b.WriteString("var azurermResourceForARMType = map[string]string{\n")
	for _, e := range clean {
		b.WriteString(fmt.Sprintf("\t%q: %q,\n", e.armType, e.name))
	}
	b.WriteString("}\n")
	if err := os.WriteFile(*outPath, []byte(b.String()), 0o644); err != nil {
		fatalf("write out: %v", err)
	}

	// Emit report.
	var r strings.Builder
	r.WriteString("# AzureRM ARM-type reference: extraction report\n\n")
	r.WriteString(fmt.Sprintf("Generated table: %d ARM types (%d unique + %d auto-resolved).\n\n", len(clean), len(clean)-len(autoResolved), len(autoResolved)))
	r.WriteString(fmt.Sprintf("## Auto-resolved ambiguous ARM types (%d) — collapsed to a main resource or discriminated base\n\n", len(autoResolved)))
	for _, g := range autoResolved {
		r.WriteString(fmt.Sprintf("- `%s` -> **%s** _(%s)_: %s\n", g.armType, g.name, g.kind, strings.Join(g.members, ", ")))
	}
	r.WriteString(fmt.Sprintf("\n## Ambiguous ARM types (%d) — several unrelated AzureRM resources; skipped\n\n", len(ambiguous)))
	for _, a := range ambiguous {
		r.WriteString("- " + a + "\n")
	}
	r.WriteString(fmt.Sprintf("\n## Unresolved resources (%d) — Create ID not traceable to a go-azure-sdk fmtString; skipped\n\n", len(skips)))
	sort.Slice(skips, func(i, j int) bool { return skips[i].name < skips[j].name })
	for _, s := range skips {
		r.WriteString(fmt.Sprintf("- `%s` (%s): %s\n", s.name, s.service, s.reason))
	}
	if err := os.WriteFile(*reportPath, []byte(r.String()), 0o644); err != nil {
		fatalf("write report: %v", err)
	}

	fmt.Fprintf(os.Stderr, "wrote %s (%d entries: %d unique + %d auto-resolved), %s (%d ambiguous, %d unresolved)\n",
		*outPath, len(clean), len(clean)-len(autoResolved), len(autoResolved), *reportPath, len(ambiguous), len(skips))
}

// resolveAmbiguous collapses a set of AzureRM resource names that share one ARM
// type to a single azapi noun when the members form an obvious family. It returns
// the chosen base name, a human-readable kind, and ok=false when no pattern fits.
//
// Two patterns are recognised, tried in order:
//
//  1. main resource: one member is a whole-word prefix of every other member, so
//     it is the parent and the others are its sub-resources
//     (azurerm_api_management is the parent of azurerm_api_management_policy, ...).
//  2. discriminated base: every member equals a common word-prefix plus a
//     non-empty distinguishing suffix (of any word length), i.e. the ARM type has
//     variant resources (azurerm_api_management_identity_provider_{aad,google,...},
//     azurerm_data_factory_credential_{service_principal,user_managed_identity}).
//     Two guards prevent over-collapsing: the prefix must carry more than the bare
//     "azurerm" token, and it must not already name another resource (taken) — the
//     latter keeps .../dataflows ambiguous, since its members share only
//     azurerm_data_factory, the name of the parent Microsoft.DataFactory/factories.
func resolveAmbiguous(names []string, taken map[string]bool) (base, kind string, ok bool) {
	for _, cand := range names {
		prefixOfAll := true
		for _, other := range names {
			if other == cand {
				continue
			}
			if !strings.HasPrefix(other, cand+"_") {
				prefixOfAll = false
				break
			}
		}
		if prefixOfAll {
			return cand, "main resource", true
		}
	}

	prefix := commonWordPrefix(names)
	prefixWords := len(strings.Split(prefix, "_"))
	if prefix != "" && prefixWords >= 2 && !taken[prefix] {
		allSuffixed := true
		for _, n := range names {
			if !strings.HasPrefix(n, prefix+"_") || len(n) <= len(prefix)+1 {
				allSuffixed = false
				break
			}
		}
		if allSuffixed {
			return prefix, "discriminated base", true
		}
	}
	return "", "", false
}

// commonWordPrefix returns the longest shared prefix of the names measured in
// whole underscore-delimited words (not characters), so "a_bc" and "a_bd" share
// "a", not "a_b".
func commonWordPrefix(names []string) string {
	if len(names) == 0 {
		return ""
	}
	parts := strings.Split(names[0], "_")
	for _, n := range names[1:] {
		w := strings.Split(n, "_")
		i := 0
		for i < len(parts) && i < len(w) && parts[i] == w[i] {
			i++
		}
		parts = parts[:i]
		if len(parts) == 0 {
			return ""
		}
	}
	return strings.Join(parts, "_")
}

// --- SDK fmtString index -----------------------------------------------------

type sdkIndexT struct {
	// key: importPath + "." + constructorFuncName -> ARM type
	byFunc map[string]string
	pkgs   map[string]bool
}

func (s *sdkIndexT) count() int    { return len(s.byFunc) }
func (s *sdkIndexT) pkgCount() int { return len(s.pkgs) }

var (
	reNewID     = regexp.MustCompile(`^func (New[A-Za-z0-9]+ID)\(`)
	reFmtString = regexp.MustCompile(`fmtString := "(/subscriptions/[^"]*/providers/[^"]+)"`)
	reProviders = regexp.MustCompile(`/providers/(.+)$`)
)

// buildSDKIndex walks the vendored go-azure-sdk resource-manager packages and the
// go-azure-helpers commonids package, mapping each New<Thing>ID constructor to the
// ARM type encoded in its package's fmtString.
func buildSDKIndex(azurermPath string) (*sdkIndexT, error) {
	idx := &sdkIndexT{byFunc: map[string]string{}, pkgs: map[string]bool{}}
	roots := []string{
		filepath.Join(azurermPath, "vendor", "github.com", "hashicorp", "go-azure-sdk", "resource-manager"),
		filepath.Join(azurermPath, "vendor", "github.com", "hashicorp", "go-azure-helpers", "resourcemanager", "commonids"),
	}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // best-effort: skip unreadable
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			armType, ctors := scanSDKFile(path)
			if armType == "" || len(ctors) == 0 {
				return nil
			}
			importPath := vendorImportPath(path)
			if importPath == "" {
				return nil
			}
			idx.pkgs[importPath] = true
			for _, c := range ctors {
				idx.byFunc[importPath+"."+c] = armType
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return idx, nil
}

// scanSDKFile returns the ARM type from the file's fmtString and the New<Thing>ID
// constructor names it defines. A go-azure-sdk id_*.go file defines exactly one ID
// type: its ID() method holds the fmtString and its constructor is New<Thing>ID.
func scanSDKFile(path string) (armType string, ctors []string) {
	f, err := os.Open(path)
	if err != nil {
		return "", nil
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if m := reNewID.FindStringSubmatch(line); m != nil {
			ctors = append(ctors, m[1])
			continue
		}
		if armType == "" {
			if m := reFmtString.FindStringSubmatch(line); m != nil {
				armType = armTypeFromFmt(m[1])
			}
		}
	}
	return armType, ctors
}

// armTypeFromFmt derives the ARM type from an ID format string. Everything after
// "/providers/" alternates literal type segments and %s placeholders (the names);
// the ARM type is the provider namespace plus each literal (non-%s) segment.
//
//	/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Kusto/clusters/%s/databases/%s
//	-> Microsoft.Kusto/clusters/databases
func armTypeFromFmt(fmtString string) string {
	m := reProviders.FindStringSubmatch(fmtString)
	if m == nil {
		return ""
	}
	segs := strings.Split(m[1], "/")
	if len(segs) == 0 {
		return ""
	}
	out := []string{segs[0]} // Microsoft.X provider
	for _, s := range segs[1:] {
		if s == "" || strings.Contains(s, "%s") {
			continue
		}
		out = append(out, s)
	}
	return strings.Join(out, "/")
}

// vendorImportPath turns a vendored file path into its Go import path (the dir).
func vendorImportPath(path string) string {
	i := strings.Index(path, "vendor"+string(filepath.Separator))
	if i < 0 {
		return ""
	}
	rel := path[i+len("vendor"+string(filepath.Separator)):]
	rel = filepath.ToSlash(filepath.Dir(rel))
	return rel
}

// --- azurerm resource extraction --------------------------------------------

type skip struct {
	name    string
	service string
	reason  string
}

var reIDAssign = regexp.MustCompile(`New[A-Za-z0-9]+ID`)

// extractResources walks each service package, enumerates the resources it
// registers (typed sdk.Resource structs and untyped SupportedResources map), and
// resolves each to its ARM type via the Create ID constructor.
func extractResources(azurermPath string, idx *sdkIndexT) (map[string]string, []skip) {
	servicesDir := filepath.Join(azurermPath, "internal", "services")
	entries, err := os.ReadDir(servicesDir)
	if err != nil {
		fatalf("read services dir: %v", err)
	}
	resources := map[string]string{}
	var skips []skip
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		svc := e.Name()
		pkgDir := filepath.Join(servicesDir, svc)
		p := parsePackage(pkgDir)
		if p == nil {
			continue
		}
		for _, res := range p.resources() {
			armType, reason := p.resolveARMType(res, idx)
			if armType == "" {
				skips = append(skips, skip{name: res.name, service: svc, reason: reason})
				continue
			}
			resources[res.name] = armType
		}
	}
	return resources, skips
}

// pkg holds a parsed azurerm service package.
type pkg struct {
	fset    *token.FileSet
	files   []*ast.File
	fileOf  map[ast.Node]*ast.File
	funcs   map[string]*ast.FuncDecl            // top-level func name -> decl
	methods map[string]map[string]*ast.FuncDecl // struct -> method name -> decl
	imports map[*ast.File]map[string]string     // file -> alias -> import path
	resType map[string]string                   // struct -> ResourceType() literal
}

func parsePackage(dir string) *pkg {
	fset := token.NewFileSet()
	des, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	p := &pkg{
		fset:    fset,
		fileOf:  map[ast.Node]*ast.File{},
		funcs:   map[string]*ast.FuncDecl{},
		methods: map[string]map[string]*ast.FuncDecl{},
		imports: map[*ast.File]map[string]string{},
		resType: map[string]string{},
	}
	for _, de := range des {
		if de.IsDir() || !strings.HasSuffix(de.Name(), ".go") || strings.HasSuffix(de.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, de.Name()), nil, 0)
		if err != nil {
			continue // best-effort
		}
		p.files = append(p.files, f)
		p.imports[f] = fileImports(f)
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			p.fileOf[fn] = f
			if fn.Recv == nil {
				p.funcs[fn.Name.Name] = fn
				continue
			}
			recv := recvType(fn.Recv)
			if recv == "" {
				continue
			}
			if p.methods[recv] == nil {
				p.methods[recv] = map[string]*ast.FuncDecl{}
			}
			p.methods[recv][fn.Name.Name] = fn
			if fn.Name.Name == "ResourceType" {
				if lit := returnedStringLit(fn); lit != "" {
					p.resType[recv] = lit
				}
			}
		}
	}
	return p
}

type resourceRef struct {
	name       string // azurerm_x
	structName string // typed: struct; "" for untyped
	ctorFunc   string // untyped: schema constructor func; "" for typed
}

// resources enumerates registered resources from registration.go decls: typed
// structs from Resources() and untyped names from SupportedResources().
func (p *pkg) resources() []resourceRef {
	var out []resourceRef
	for _, f := range p.files {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil {
				continue
			}
			switch fn.Name.Name {
			case "Resources":
				out = append(out, p.typedResources(fn)...)
			case "SupportedResources":
				out = append(out, p.untypedResources(fn)...)
			}
		}
	}
	return out
}

// typedResources reads the []sdk.Resource{ StructName{}, ... } slice literal.
func (p *pkg) typedResources(fn *ast.FuncDecl) []resourceRef {
	var out []resourceRef
	ast.Inspect(fn, func(n ast.Node) bool {
		cl, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		for _, el := range cl.Elts {
			inner, ok := el.(*ast.CompositeLit)
			if !ok {
				continue
			}
			if id, ok := inner.Type.(*ast.Ident); ok {
				name := p.resType[id.Name]
				if name == "" {
					continue
				}
				out = append(out, resourceRef{name: name, structName: id.Name})
			}
		}
		return true
	})
	return out
}

// untypedResources reads the map[string]*pluginsdk.Resource{ "azurerm_x": ctor() }
// literal, recording the schema constructor func name for each resource name.
func (p *pkg) untypedResources(fn *ast.FuncDecl) []resourceRef {
	var out []resourceRef
	ast.Inspect(fn, func(n ast.Node) bool {
		cl, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		for _, el := range cl.Elts {
			kv, ok := el.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key, ok := kv.Key.(*ast.BasicLit)
			if !ok || key.Kind != token.STRING {
				continue
			}
			name, err := strconv.Unquote(key.Value)
			if err != nil || !strings.HasPrefix(name, "azurerm_") {
				continue
			}
			ctor := ""
			if call, ok := kv.Value.(*ast.CallExpr); ok {
				if id, ok := call.Fun.(*ast.Ident); ok {
					ctor = id.Name
				}
			}
			out = append(out, resourceRef{name: name, ctorFunc: ctor})
		}
		return true
	})
	return out
}

// resolveARMType finds the resource's Create func, extracts id := <alias>.New*ID(,
// resolves the alias to an SDK import path, and looks up the ARM type.
func (p *pkg) resolveARMType(res resourceRef, idx *sdkIndexT) (armType, reason string) {
	var createFn *ast.FuncDecl
	if res.structName != "" {
		createFn = p.methods[res.structName]["Create"]
		if createFn == nil {
			return "", "no Create method"
		}
	} else {
		// Untyped: the CRUD func lives beside the schema constructor. Try common
		// naming, then fall back to the constructor func's own body.
		base := res.ctorFunc
		for _, suffix := range []string{"Create", "CreateUpdate", "CreateFunc"} {
			if fn := p.funcs[base+suffix]; fn != nil {
				createFn = fn
				break
			}
		}
		if createFn == nil {
			createFn = p.funcs[base]
		}
		if createFn == nil {
			return "", "no Create func for " + base
		}
	}
	alias, ctor := findIDConstructor(createFn)
	if ctor == "" {
		return "", "no id := New*ID constructor in Create"
	}
	file := p.fileOf[createFn]
	importPath := p.imports[file][alias]
	if importPath == "" {
		return "", "unresolved alias " + alias
	}
	if arm, ok := idx.byFunc[importPath+"."+ctor]; ok {
		return arm, ""
	}
	return "", fmt.Sprintf("no SDK fmtString for %s.%s", importPath, ctor)
}

// findIDConstructor finds the first `id := <alias>.New<Thing>ID(...)` assignment
// in the func body and returns the alias and constructor name.
func findIDConstructor(fn *ast.FuncDecl) (alias, ctor string) {
	if fn.Body == nil {
		return "", ""
	}
	var foundAlias, foundCtor string
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if foundCtor != "" {
			return false
		}
		as, ok := n.(*ast.AssignStmt)
		if !ok || len(as.Lhs) != 1 || len(as.Rhs) != 1 {
			return true
		}
		lhs, ok := as.Lhs[0].(*ast.Ident)
		if !ok || lhs.Name != "id" {
			return true
		}
		call, ok := as.Rhs[0].(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkgIdent, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		if !reIDAssign.MatchString(sel.Sel.Name) {
			return true
		}
		foundAlias, foundCtor = pkgIdent.Name, sel.Sel.Name
		return false
	})
	return foundAlias, foundCtor
}

// --- ast helpers -------------------------------------------------------------

func fileImports(f *ast.File) map[string]string {
	m := map[string]string{}
	for _, imp := range f.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		alias := ""
		if imp.Name != nil {
			alias = imp.Name.Name
		} else {
			alias = path[strings.LastIndex(path, "/")+1:]
		}
		m[alias] = path
	}
	return m
}

func recvType(fl *ast.FieldList) string {
	if fl == nil || len(fl.List) == 0 {
		return ""
	}
	switch t := fl.List[0].Type.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return id.Name
		}
	}
	return ""
}

func returnedStringLit(fn *ast.FuncDecl) string {
	if fn.Body == nil {
		return ""
	}
	var lit string
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok || len(ret.Results) != 1 {
			return true
		}
		if bl, ok := ret.Results[0].(*ast.BasicLit); ok && bl.Kind == token.STRING {
			if s, err := strconv.Unquote(bl.Value); err == nil {
				lit = s
			}
		}
		return true
	})
	return lit
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "azurerm-armtypes: "+format+"\n", args...)
	os.Exit(1)
}
