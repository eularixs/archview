// Package classify maps a Go import path to an architectural layer
// (controller / service / repository / other) and a best-effort module name,
// using folder- and naming-convention heuristics.
//
// It is intentionally convention-driven so it works across modular-MVC and
// hexagonal layouts without per-project annotation. A Config can extend the
// keyword sets.
package classify

import (
	"strings"

	"github.com/eularixs/archview/graph"
)

// layerKeywords maps a canonical layer to the path/package keywords that imply
// it. Matched case-insensitively, either as a whole path segment or as a suffix
// of one (e.g. "user_service" -> service, module "user").
var layerKeywords = map[string][]string{
	// Inbound / driving side: MVC controllers + hexagonal transport/interface adapters.
	graph.LayerController: {"controller", "controllers", "handler", "handlers", "delivery", "rest", "transport", "grpc", "graphql", "web", "interface", "interfaces"},
	// Application core: MVC services + hexagonal use cases / clean-arch core.
	graph.LayerService: {"service", "services", "usecase", "usecases", "use_case", "interactor", "interactors", "application", "logic", "core"},
	// Outbound / driven side: MVC repositories + hexagonal persistence/gateway adapters.
	graph.LayerRepository: {"repository", "repositories", "repo", "repos", "store", "stores", "dao", "persistence", "gateway", "postgres", "postgresql", "mysql", "mongo", "mongodb", "sqlite", "supabase"},
}

// generic path segments that are containers, not module names. Includes
// hexagonal structural dirs so the module resolves to the bounded-context name
// (e.g. catalog/adapter/rest -> module "catalog").
var generic = map[string]bool{
	"modules": true, "module": true, "internal": true, "app": true,
	"pkg": true, "src": true, "cmd": true, "features": true, "feature": true,
	"components": true, "component": true, "domain": true,
	"adapter": true, "adapters": true, "port": true, "ports": true,
	"inbound": true, "outbound": true, "driving": true, "driven": true,
	"infra": true, "infrastructure": true,
}

// Config allows extending the default keyword sets (optional override).
type Config struct {
	// Extra maps a canonical layer (graph.LayerController/Service/Repository)
	// to additional keywords to recognize.
	Extra map[string][]string
}

// Classifier classifies import paths into (layer, module).
type Classifier struct {
	keywords map[string][]string
}

// New builds a Classifier from the default keyword sets plus any config extras.
func New(cfg *Config) *Classifier {
	kw := make(map[string][]string, len(layerKeywords))
	for layer, ks := range layerKeywords {
		kw[layer] = append([]string(nil), ks...)
	}
	if cfg != nil {
		for layer, extra := range cfg.Extra {
			kw[layer] = append(kw[layer], extra...)
		}
	}
	return &Classifier{keywords: kw}
}

// Classify returns the layer and module for an import path. Layer falls back to
// graph.LayerOther when nothing matches; module is "" when undeterminable.
//
// The module is the bounded-context name. For feature-first layouts
// (user/service) it is the segment before the layer keyword; for layer-first
// layouts (core/user, interface/user, handlers/user) it is the segment after.
func (c *Classifier) Classify(pkgPath string) (layer, module string) {
	segs := strings.Split(pkgPath, "/")
	// Search from the most specific (tail) segment backwards.
	for i := len(segs) - 1; i >= 0; i-- {
		s := strings.ToLower(segs[i])
		if l, base, ok := c.matchLayer(s); ok {
			if base != "" {
				return l, base
			}
			// Prefer feature-first (module before the layer); fall back to
			// layer-first (module after).
			if m := moduleFromPrev(segs, i); m != "" {
				return l, m
			}
			return l, moduleFromNext(segs, i)
		}
	}
	return graph.LayerOther, moduleFromPrev(segs, len(segs))
}

// matchLayer reports whether a single segment denotes a layer. If the keyword
// is a suffix of a longer segment (e.g. "user_service"), base holds the trimmed
// remainder ("user") to use as the module.
func (c *Classifier) matchLayer(seg string) (layer, base string, ok bool) {
	for layer, ks := range c.keywords {
		for _, k := range ks {
			if seg == k {
				return layer, "", true
			}
			if strings.HasSuffix(seg, k) && len(seg) > len(k) {
				base := strings.Trim(seg[:len(seg)-len(k)], "_-.")
				return layer, base, true
			}
		}
	}
	return "", "", false
}

// moduleFromPrev walks back from index i, returning the first non-generic
// segment as the module name, or "" if none.
func moduleFromPrev(segs []string, i int) string {
	for j := i - 1; j >= 0; j-- {
		s := strings.ToLower(segs[j])
		if s == "" || generic[s] {
			continue
		}
		return segs[j]
	}
	return ""
}

// moduleFromNext walks forward from index i, returning the first non-generic
// segment as the module name, or "" if none. Used for layer-first layouts where
// the feature follows the layer dir (e.g. core/user -> user).
func moduleFromNext(segs []string, i int) string {
	for j := i + 1; j < len(segs); j++ {
		s := strings.ToLower(segs[j])
		if s == "" || generic[s] {
			continue
		}
		return segs[j]
	}
	return ""
}
