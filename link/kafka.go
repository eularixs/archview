package link

import (
	"go/ast"
	"go/constant"
	"go/types"
	"strings"

	"github.com/eularixs/archview/analyzer"
	"github.com/eularixs/archview/route"
)

// kafkaExtractor stitches message-queue events: a subscriber (kafka.ReaderConfig,
// surfaced as a SUB endpoint) exposes "kafka:<topic>", and a publisher
// (kafka.Writer{Topic: ...}) consumes the same key. The stitched edge runs
// publisher -> subscriber, the direction the event flows.
type kafkaExtractor struct{}

func (kafkaExtractor) Name() string { return "kafka" }

// Inbound reuses the SUB routes (kafka subscriptions), keyed by topic.
func (kafkaExtractor) Inbound(res *analyzer.Result) []Link {
	var out []Link
	for _, r := range route.Extract(res.Pkgs, route.Default()) {
		if r.Method != "SUB" {
			continue
		}
		f := res.FuncFor(r.Handler)
		if f == nil {
			continue
		}
		out = append(out, Link{Key: "kafka:" + r.Path, Func: f.SSA, Kind: "kafka"})
	}
	return out
}

// Outbound finds kafka.Writer{Topic: ...} literals; the enclosing function is the
// publisher and the topic is the key.
func (kafkaExtractor) Outbound(res *analyzer.Result) []Link {
	var out []Link
	for _, pkg := range res.Pkgs {
		info := pkg.TypesInfo
		if info == nil {
			continue
		}
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				fd, ok := decl.(*ast.FuncDecl)
				if !ok || fd.Body == nil {
					continue
				}
				obj, _ := info.Defs[fd.Name].(*types.Func)
				f := res.FuncFor(obj)
				if f == nil {
					continue
				}
				ast.Inspect(fd.Body, func(n ast.Node) bool {
					cl, ok := n.(*ast.CompositeLit)
					if !ok || !isKafkaNamed(info.TypeOf(cl), "Writer") {
						return true
					}
					if topic := kafkaTopic(info, cl); topic != "" {
						out = append(out, Link{Key: "kafka:" + topic, Func: f.SSA, Kind: "kafka"})
					}
					return true
				})
			}
		}
	}
	return out
}

func isKafkaNamed(t types.Type, name string) bool {
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	n, ok := t.(*types.Named)
	if !ok || n.Obj() == nil || n.Obj().Name() != name {
		return false
	}
	pkg := n.Obj().Pkg()
	return pkg != nil && strings.Contains(pkg.Path(), "kafka")
}

func kafkaTopic(info *types.Info, cl *ast.CompositeLit) string {
	for _, e := range cl.Elts {
		kv, ok := e.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		if id, ok := kv.Key.(*ast.Ident); !ok || id.Name != "Topic" {
			continue
		}
		if tv, ok := info.Types[kv.Value]; ok && tv.Value != nil && tv.Value.Kind() == constant.String {
			return constant.StringVal(tv.Value)
		}
	}
	return ""
}
