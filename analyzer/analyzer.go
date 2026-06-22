// Package analyzer loads a Go module from source, builds its SSA form and a
// CHA (Class Hierarchy Analysis) call graph, and exposes the project's own
// functions with their source positions. CHA resolves interface dispatch to all
// implementing types, which is accurate for the common single-implementation
// layered case (controller -> service -> repository through interfaces).
package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"regexp"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/cha"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

// genMarker matches the standard Go "generated file" header. Functions in such
// files (protoc, sqlc, mockgen, ent, …) are skipped so generated boilerplate
// doesn't clutter the graph.
var genMarker = regexp.MustCompile(`^// Code generated .* DO NOT EDIT\.$`)

func isGeneratedFile(f *ast.File) bool {
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			if genMarker.MatchString(c.Text) {
				return true
			}
		}
	}
	return false
}

// Func is a project function/method discovered during analysis.
type Func struct {
	SSA  *ssa.Function
	Pkg  string // import path
	Recv string // receiver type without package qualifier, e.g. "*UserController" ("" for free functions)
	Name string // method/function name, e.g. "GetUser"
	File string // absolute source file
	Line int
	Col  int
}

// Display returns a human label like "(*UserController).GetUser" or "GetUser".
func (f *Func) Display() string {
	if f.Recv != "" {
		return fmt.Sprintf("(%s).%s", f.Recv, f.Name)
	}
	return f.Name
}

// Result is the analyzed program.
type Result struct {
	Module    string // main module path (may be "")
	Fset      *token.FileSet
	Prog      *ssa.Program
	Pkgs      []*packages.Package
	CallGraph *callgraph.Graph

	// Funcs holds only project-owned functions, keyed by their SSA function.
	Funcs map[*ssa.Function]*Func

	inProject map[string]bool // import paths owned by the project
}

// Load analyzes the Go module rooted at dir ("." for cwd).
func Load(dir string) (*Result, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedDeps | packages.NeedTypes |
			packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedModule,
		Dir:   dir,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no Go packages found under %q", dir)
	}

	// Project-owned import paths = the matched packages (deps live behind their
	// .Imports and are excluded from this set).
	inProject := make(map[string]bool, len(pkgs))
	var module string
	for _, p := range pkgs {
		if p.PkgPath != "" {
			inProject[p.PkgPath] = true
		}
		if module == "" && p.Module != nil {
			module = p.Module.Path
		}
	}

	prog, _ := ssautil.AllPackages(pkgs, ssa.InstantiateGenerics)
	prog.Build()
	cg := cha.CallGraph(prog)
	cg.DeleteSyntheticNodes()

	// Generated source files (their functions are excluded below).
	generated := map[string]bool{}
	for _, p := range pkgs {
		for _, f := range p.Syntax {
			if isGeneratedFile(f) {
				abs := prog.Fset.Position(f.Pos()).Filename
				if a, err := filepath.Abs(abs); err == nil {
					abs = a
				}
				generated[abs] = true
			}
		}
	}

	res := &Result{
		Module:    module,
		Fset:      prog.Fset,
		Prog:      prog,
		Pkgs:      pkgs,
		CallGraph: cg,
		Funcs:     map[*ssa.Function]*Func{},
		inProject: inProject,
	}

	for fn := range ssautil.AllFunctions(prog) {
		if !res.owns(fn) {
			continue
		}
		f := res.toFunc(fn)
		if generated[f.File] {
			continue // skip generated boilerplate (protoc, sqlc, …)
		}
		res.Funcs[fn] = f
	}
	return res, nil
}

// InProject reports whether an import path belongs to the analyzed project.
func (r *Result) InProject(pkgPath string) bool { return r.inProject[pkgPath] }

// owns reports whether an SSA function is a real project function we should
// surface (has a package, a position, and lives in a project package).
func (r *Result) owns(fn *ssa.Function) bool {
	if fn == nil || fn.Pkg == nil || fn.Pkg.Pkg == nil {
		return false
	}
	if !fn.Pos().IsValid() {
		return false
	}
	return r.inProject[fn.Pkg.Pkg.Path()]
}

// FuncFor resolves a types.Func (e.g. a route handler) to its project Func, or
// nil if it isn't a known project function.
func (r *Result) FuncFor(obj *types.Func) *Func {
	if obj == nil {
		return nil
	}
	fn := r.Prog.FuncValue(obj)
	if fn == nil {
		return nil
	}
	return r.Funcs[fn]
}

func (r *Result) toFunc(fn *ssa.Function) *Func {
	pos := r.Fset.Position(fn.Pos())
	abs := pos.Filename
	if a, err := filepath.Abs(abs); err == nil {
		abs = a
	}
	recv := ""
	if sig, ok := fn.Type().(*types.Signature); ok && sig.Recv() != nil {
		recv = types.TypeString(sig.Recv().Type(), func(p *types.Package) string { return "" })
	}
	return &Func{
		SSA:  fn,
		Pkg:  fn.Pkg.Pkg.Path(),
		Recv: recv,
		Name: fn.Name(),
		File: abs,
		Line: pos.Line,
		Col:  pos.Column,
	}
}
