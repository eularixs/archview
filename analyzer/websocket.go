package analyzer

import (
	"strings"

	"golang.org/x/tools/go/ssa"
)

// WSUpgraders returns the project functions that perform a WebSocket upgrade —
// i.e. they call a websocket package's Upgrade (gorilla/websocket) or Accept
// (coder/nhooyr websocket). These are the handlers behind a WebSocket endpoint,
// so the builder can label their route "WS".
func (r *Result) WSUpgraders() map[*ssa.Function]bool {
	out := map[*ssa.Function]bool{}
	for fn := range r.Funcs {
		for _, b := range fn.Blocks {
			for _, instr := range b.Instrs {
				cc := callCommon(instr)
				if cc == nil {
					continue
				}
				var pkg, name string
				if callee := cc.StaticCallee(); callee != nil && callee.Pkg != nil && callee.Pkg.Pkg != nil {
					pkg, name = callee.Pkg.Pkg.Path(), callee.Name()
				} else if cc.IsInvoke() && cc.Method != nil && cc.Method.Pkg() != nil {
					pkg, name = cc.Method.Pkg().Path(), cc.Method.Name()
				}
				if isWSUpgrade(pkg, name) {
					out[fn] = true
				}
			}
		}
	}
	return out
}

func isWSUpgrade(pkg, name string) bool {
	if !strings.Contains(strings.ToLower(pkg), "websocket") {
		return false
	}
	return name == "Upgrade" || name == "Accept"
}
