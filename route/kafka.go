package route

import (
	"go/ast"
	"go/constant"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// kafkaExtractor detects message-queue subscribers (segmentio/kafka-go) as
// inbound entrypoints: a kafka.ReaderConfig{Topic: ...} is a subscription, so
// its enclosing function is the consumer handler and the topic is the route
// path. This surfaces pure-consumer services (no HTTP/gRPC endpoint) that would
// otherwise be pruned, and provides the inbound side of queue stitching.
type kafkaExtractor struct{}

func (kafkaExtractor) Name() string { return "kafka" }

func (kafkaExtractor) Match(pkg *packages.Package) bool { return len(kafkaSubs(pkg)) > 0 }

func (kafkaExtractor) Extract(pkg *packages.Package) []Route { return kafkaSubs(pkg) }

func kafkaSubs(pkg *packages.Package) []Route {
	info := pkg.TypesInfo
	if info == nil {
		return nil
	}
	var out []Route
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				continue
			}
			handler, _ := info.Defs[fd.Name].(*types.Func)
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				cl, ok := n.(*ast.CompositeLit)
				if !ok || !isKafkaNamed(info.TypeOf(cl), "ReaderConfig") {
					return true
				}
				topic := kafkaTopicField(info, cl)
				if topic == "" {
					return true
				}
				out = append(out, Route{Method: "SUB", Path: topic, Handler: handler, Pos: cl.Pos()})
				return true
			})
		}
	}
	return out
}

// isKafkaNamed reports whether t is a named type called name from a kafka package.
func isKafkaNamed(t types.Type, name string) bool {
	n := namedType(t)
	if n == nil || n.Obj() == nil {
		return false
	}
	if n.Obj().Name() != name {
		return false
	}
	pkg := n.Obj().Pkg()
	return pkg != nil && strings.Contains(pkg.Path(), "kafka")
}

// kafkaTopicField returns the constant string value of the literal's Topic field.
func kafkaTopicField(info *types.Info, cl *ast.CompositeLit) string {
	for _, e := range cl.Elts {
		kv, ok := e.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		id, ok := kv.Key.(*ast.Ident)
		if !ok || id.Name != "Topic" {
			continue
		}
		if tv, ok := info.Types[kv.Value]; ok && tv.Value != nil && tv.Value.Kind() == constant.String {
			return constant.StringVal(tv.Value)
		}
	}
	return ""
}
