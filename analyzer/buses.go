package analyzer

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/ssa"
)

// handlerMethodNames are the primary method names a mediator handler exposes.
var handlerMethodNames = map[string]bool{
	"Handle": true, "Execute": true, "Run": true, "Serve": true, "Process": true, "On": true,
}

// Dispatch is a resolved mediator dispatch: a caller sends a message through a
// bus and the message routes to one or more concrete handlers.
type Dispatch struct {
	Caller   *ssa.Function
	Message  string
	Handlers []*ssa.Function
}

// BusInfo is the result of bus detection: the bus methods (used as call-graph
// barriers so CHA's over-approximation through the bus is dropped) and the
// precise dispatch edges recovered from registration sites.
type BusInfo struct {
	BusMethods map[*ssa.Function]bool
	Dispatches []Dispatch
}

// Buses detects command/query/event mediators. It learns bus types and their
// message->handler routing from registration calls (e.g.
// bus.Register(Cmd{}.Name(), NewCmdHandler(...))), then resolves dispatch sites
// (calls on a bus type passing a routed message) to the concrete handlers.
//
// This recovers the routing a static call graph cannot: the bus stores handlers
// in a map keyed at runtime, so CHA would otherwise fan every dispatch out to
// every handler that implements the handler interface.
func (r *Result) Buses() *BusInfo {
	info := &BusInfo{BusMethods: map[*ssa.Function]bool{}}

	busTypes := map[*types.TypeName]bool{}
	routing := map[*types.TypeName][]*ssa.Function{} // message type -> handlers

	// Pass 1: registration calls — recv.Method(key, handler), 2 args.
	for _, p := range r.Pkgs {
		info0 := p.TypesInfo
		if info0 == nil {
			continue
		}
		for _, file := range p.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok || len(call.Args) != 2 {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				handlerType := namedOf(info0.TypeOf(call.Args[1]))
				if handlerType == nil {
					return true
				}
				hf := r.methodFunc(handlerType, handlerMethodNames)
				if hf == nil || r.Funcs[hf] == nil {
					return true
				}
				busType := namedOf(info0.TypeOf(sel.X))
				if busType == nil {
					return true
				}
				busTypes[busType.Obj()] = true

				// Message type from the key arg: Msg{}.Name() -> typeof(Msg{}).
				if inner, ok := call.Args[0].(*ast.CallExpr); ok {
					if isel, ok := inner.Fun.(*ast.SelectorExpr); ok {
						if mt := namedOf(info0.TypeOf(isel.X)); mt != nil {
							routing[mt.Obj()] = append(routing[mt.Obj()], hf)
						}
					}
				}
				return true
			})
		}
	}
	if len(busTypes) == 0 {
		return info
	}

	// Bus methods = every project method whose receiver is a bus type.
	for fn := range r.Funcs {
		if tn := recvTypeName(fn); tn != nil && busTypes[tn] {
			info.BusMethods[fn] = true
		}
	}

	// Pass 2: dispatch sites — a static call to a bus method passing a routed
	// message value.
	for fn := range r.Funcs {
		for _, b := range fn.Blocks {
			for _, instr := range b.Instrs {
				cc := callCommon(instr)
				if cc == nil {
					continue
				}
				callee := cc.StaticCallee()
				if callee == nil || !info.BusMethods[callee] {
					continue
				}
				for _, arg := range cc.Args {
					// The message is boxed into the bus's interface parameter;
					// unwrap to recover the concrete command/event type.
					if mi, ok := arg.(*ssa.MakeInterface); ok {
						arg = mi.X
					}
					mt := namedOf(arg.Type())
					if mt == nil {
						continue
					}
					if hs := routing[mt.Obj()]; len(hs) > 0 {
						info.Dispatches = append(info.Dispatches, Dispatch{
							Caller:   fn,
							Message:  mt.Obj().Name(),
							Handlers: hs,
						})
					}
				}
			}
		}
	}
	return info
}

// methodFunc returns the project SSA function for the first method of named
// whose name is in names. The value receiver is tried before the pointer so a
// value-receiver method resolves to its real function rather than a synthesized
// pointer wrapper (which has no source position and isn't in r.Funcs).
func (r *Result) methodFunc(named *types.Named, names map[string]bool) *ssa.Function {
	for _, recv := range []types.Type{named, types.NewPointer(named)} {
		ms := types.NewMethodSet(recv)
		for i := 0; i < ms.Len(); i++ {
			sel := ms.At(i)
			if !names[sel.Obj().Name()] {
				continue
			}
			if fn := r.Prog.MethodValue(sel); fn != nil && r.Funcs[fn] != nil {
				return fn
			}
		}
	}
	return nil
}

// namedOf unwraps a pointer and returns the underlying named type, or nil.
func namedOf(t types.Type) *types.Named {
	if t == nil {
		return nil
	}
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	if n, ok := t.(*types.Named); ok {
		return n
	}
	return nil
}

// recvTypeName returns the receiver's named type object for a method, or nil.
func recvTypeName(fn *ssa.Function) *types.TypeName {
	if fn.Signature == nil || fn.Signature.Recv() == nil {
		return nil
	}
	if n := namedOf(fn.Signature.Recv().Type()); n != nil {
		return n.Obj()
	}
	return nil
}

// callCommon extracts the call payload from a call-like instruction.
func callCommon(instr ssa.Instruction) *ssa.CallCommon {
	switch c := instr.(type) {
	case *ssa.Call:
		return &c.Call
	case *ssa.Defer:
		return &c.Call
	case *ssa.Go:
		return &c.Call
	}
	return nil
}
