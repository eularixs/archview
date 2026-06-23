// Package link recovers inter-service integration edges that a single-process
// call graph cannot: in a microservice repo, one service exposes a contract
// (gRPC method, HTTP route, queue topic) and another consumes it over the wire.
// Every transport reduces to the same shape — a producer emits a key, a consumer
// references it — so adding a transport (Kafka, NATS, HTTP) is one Extractor,
// with the stitching core unchanged.
package link

import (
	"github.com/eularixs/archview/analyzer"
	"golang.org/x/tools/go/ssa"
)

// Link is one keyed integration point: a function that exposes (inbound) or
// consumes (outbound) a transport key such as "grpc:UserService/GetUser".
type Link struct {
	Key  string
	Func *ssa.Function
	Kind string // "grpc" | "http" | "kafka" | ...
}

// Extractor detects one transport's integration points across the whole module.
type Extractor interface {
	Name() string
	Inbound(res *analyzer.Result) []Link  // contracts this code exposes (server / subscriber)
	Outbound(res *analyzer.Result) []Link // contracts this code consumes (client / publisher)
}

// Default returns the built-in extractors. gRPC first; others are additive.
func Default() []Extractor {
	return []Extractor{grpcExtractor{}}
}

// Cross is a stitched cross-service call: From consumes the key that To exposes.
type Cross struct {
	From *ssa.Function
	To   *ssa.Function
	Key  string
	Kind string
}

// Stitch matches each outbound link to the inbound links sharing its key,
// yielding a cross-service edge per (consumer, producer) pair.
func Stitch(res *analyzer.Result, extractors []Extractor) []Cross {
	inbound := map[string][]Link{}
	var outbound []Link
	for _, e := range extractors {
		for _, l := range e.Inbound(res) {
			if l.Func != nil {
				inbound[l.Key] = append(inbound[l.Key], l)
			}
		}
		outbound = append(outbound, e.Outbound(res)...)
	}
	var crosses []Cross
	for _, o := range outbound {
		if o.Func == nil {
			continue
		}
		for _, in := range inbound[o.Key] {
			if in.Func == o.Func {
				continue
			}
			crosses = append(crosses, Cross{From: o.Func, To: in.Func, Key: o.Key, Kind: o.Kind})
		}
	}
	return crosses
}

// callCommon extracts the call payload from any call-like SSA instruction.
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
