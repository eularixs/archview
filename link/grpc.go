package link

import (
	"go/types"
	"strings"

	"github.com/eularixs/archview/analyzer"
	"github.com/eularixs/archview/route"
)

// grpcExtractor stitches gRPC: the server side (Register<Svc>Server, surfaced by
// the route extractors as RPC endpoints) exposes "grpc:<Svc>/<Method>", and a
// client invoke on a generated <Svc>Client interface consumes the same key.
type grpcExtractor struct{}

func (grpcExtractor) Name() string { return "grpc" }

// Inbound reuses the route extractors' RPC routes (gRPC + ConnectRPC), keying
// each by its proto service and method.
func (grpcExtractor) Inbound(res *analyzer.Result) []Link {
	var out []Link
	for _, r := range route.Extract(res.Pkgs, route.Default()) {
		if r.Method != "RPC" {
			continue
		}
		f := res.FuncFor(r.Handler)
		if f == nil {
			continue
		}
		out = append(out, Link{Key: "grpc:" + strings.TrimPrefix(r.Path, "/"), Func: f.SSA, Kind: "grpc"})
	}
	return out
}

// Outbound finds calls on a generated client: an interface method invoke whose
// receiver type is named "<Svc>Client". The key is "grpc:<Svc>/<Method>".
func (grpcExtractor) Outbound(res *analyzer.Result) []Link {
	var out []Link
	for fn := range res.Funcs {
		for _, b := range fn.Blocks {
			for _, instr := range b.Instrs {
				cc := callCommon(instr)
				if cc == nil || !cc.IsInvoke() || cc.Method == nil {
					continue
				}
				named, ok := cc.Value.Type().(*types.Named)
				if !ok {
					continue
				}
				svc := named.Obj().Name()
				if !strings.HasSuffix(svc, "Client") {
					continue
				}
				svc = strings.TrimSuffix(svc, "Client")
				out = append(out, Link{
					Key:  "grpc:" + svc + "/" + cc.Method.Name(),
					Func: fn,
					Kind: "grpc",
				})
			}
		}
	}
	return out
}
