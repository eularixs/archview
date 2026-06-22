package analyzer

import (
	"go/types"
	"path/filepath"

	"golang.org/x/tools/go/ssa"
)

// Port is a project-owned interface that is implemented by a project concrete
// type — i.e. a hexagonal "port". It records who uses it (callers that invoke
// its methods through the interface) and who implements it (the adapter methods
// that satisfy it). The builder decides whether a port is inbound or outbound
// from the layer of its implementers.
type Port struct {
	Name string
	Pkg  string
	File string
	Line int
	Col  int

	Methods     []string
	Callers     map[*ssa.Function]bool // project funcs that invoke a method on this interface
	ImplMethods map[*ssa.Function]bool // project methods that satisfy this interface
}

// Ports finds every project interface implemented by a project concrete type,
// resolving its callers (interface-dispatch sites) and implementer methods.
func (r *Result) Ports() []*Port {
	type ifaceInfo struct {
		obj   *types.TypeName
		named *types.Named
		iface *types.Interface
	}
	var ifaces []ifaceInfo
	var concretes []*types.Named

	for _, p := range r.Pkgs {
		if p.Types == nil {
			continue
		}
		scope := p.Types.Scope()
		for _, name := range scope.Names() {
			tn, ok := scope.Lookup(name).(*types.TypeName)
			if !ok {
				continue
			}
			named, ok := tn.Type().(*types.Named)
			if !ok {
				continue
			}
			switch u := named.Underlying().(type) {
			case *types.Interface:
				if u.NumMethods() > 0 {
					ifaces = append(ifaces, ifaceInfo{tn, named, u})
				}
			case *types.Struct:
				concretes = append(concretes, named)
			}
		}
	}

	byObj := map[*types.TypeName]*Port{}
	for _, info := range ifaces {
		methodNames := map[string]bool{}
		var methods []string
		for i := 0; i < info.iface.NumMethods(); i++ {
			m := info.iface.Method(i)
			methodNames[m.Name()] = true
			methods = append(methods, m.Name())
		}

		impl := map[*ssa.Function]bool{}
		for _, ct := range concretes {
			ptr := types.NewPointer(ct)
			if !types.Implements(ptr, info.iface) && !types.Implements(ct, info.iface) {
				continue
			}
			ms := types.NewMethodSet(ptr)
			for i := 0; i < ms.Len(); i++ {
				sel := ms.At(i)
				if !methodNames[sel.Obj().Name()] {
					continue
				}
				if fn := r.Prog.MethodValue(sel); fn != nil && r.Funcs[fn] != nil {
					impl[fn] = true
				}
			}
		}
		if len(impl) == 0 {
			continue // not implemented by any project type — nothing to draw
		}

		pos := r.Fset.Position(info.obj.Pos())
		abs := pos.Filename
		if a, err := filepath.Abs(abs); err == nil {
			abs = a
		}
		pkgPath := ""
		if info.obj.Pkg() != nil {
			pkgPath = info.obj.Pkg().Path()
		}
		byObj[info.obj] = &Port{
			Name:        info.named.Obj().Name(),
			Pkg:         pkgPath,
			File:        abs,
			Line:        pos.Line,
			Col:         pos.Column,
			Methods:     methods,
			Callers:     map[*ssa.Function]bool{},
			ImplMethods: impl,
		}
	}

	// Resolve callers: any project func with an interface-dispatch (invoke-mode)
	// call to one of our ports.
	for fn := range r.Funcs {
		for _, b := range fn.Blocks {
			for _, instr := range b.Instrs {
				var cc *ssa.CallCommon
				switch c := instr.(type) {
				case *ssa.Call:
					cc = &c.Call
				case *ssa.Defer:
					cc = &c.Call
				case *ssa.Go:
					cc = &c.Call
				}
				if cc == nil || !cc.IsInvoke() {
					continue
				}
				named, ok := cc.Value.Type().(*types.Named)
				if !ok {
					continue
				}
				if _, ok := named.Underlying().(*types.Interface); !ok {
					continue
				}
				if p := byObj[named.Obj()]; p != nil {
					p.Callers[fn] = true
				}
			}
		}
	}

	out := make([]*Port, 0, len(byObj))
	for _, p := range byObj {
		out = append(out, p)
	}
	return out
}
