# archview

Live architecture flow graph for Go backends. Mount it in `main.go`, open
`/graph` in a browser, and see your endpoints flow through
controller → service → repository — generated automatically from your source,
no annotations. Click any node to jump to its definition in your editor.

Framework-agnostic (net/http, gin, echo, gRPC, GraphQL) and pattern-aware
(modular MVC, hexagonal ports & adapters, CQRS / event buses): the arrows
follow the call graph of whatever layout you actually use.

> Status: **v0.3** — dev-live mode; net/http + gin + echo + gRPC + GraphQL;
> modular MVC, hexagonal (outbound ports), and CQRS/mediator (bus detection);
> chain-based auto-layer for naming-agnostic analysis.
> See [`docs/`](docs/) for the PRD, plan and roadmap.

## Install

```sh
go get github.com/eularixs/archview
```

## Usage

```go
mux := http.NewServeMux()

av, err := archview.New(archview.Options{
    Root:        ".",      // module dir to analyze (default ".")
    BasePath:    "/graph", // mount path (default "/graph")
    Editor:      "vscode", // click-to-source: "vscode" or "cursor"
    ShowPorts:   true,     // surface hexagonal outbound ports (default off)
    DetectBuses: true,     // recover command/query/event routing (default off)
    // AutoLayer is ON by default; set DisableAutoLayer: true for keyword-only.
})
if err != nil {
    log.Fatal(err)
}
av.Mount(mux)             // serves /graph and /graph/data

http.ListenAndServe(":8080", mux)
```

Open <http://localhost:8080/graph>. `ShowPorts` and `DetectBuses` are opt-in,
so a plain MVC graph is unchanged unless you turn them on.

Using gin (or any framework)? archview owns its own path; let your framework
handle the rest:

```go
mux := http.NewServeMux()
av.Mount(mux)
mux.Handle("/", ginEngine) // gin handles everything except /graph
```

## How it works

```
source → go/packages (AST+types) → CHA call graph
       → classify layers (folder/naming) → extract routes (per framework)
       → graph model → /graph (embedded SVG renderer)
```

archview statically analyzes the module at startup (**dev-live**): it builds an
SSA call graph, detects HTTP route registrations per framework, classifies each
function into a layer, and renders the resulting node/edge graph. Every function
node carries a `file:line:col`, so clicking it opens
`vscode://file/...` (or `cursor://...`).

dev-live needs the source tree and Go toolchain present at runtime.

## Layer classification

By convention, from the import path / package name:

| Layer       | Keywords (MVC + hexagonal) |
|-------------|----------------------------|
| controller  | `controller(s)`, `handler(s)`, `delivery`, `rest`, `transport`, `grpc`, `graphql`, `web` |
| service     | `service(s)`, `usecase(s)`, `interactor`, `application`, `logic` |
| repository  | `repository`, `repo(s)`, `store(s)`, `dao`, `persistence`, `gateway`, `postgres`, `mysql`, `mongo`, `sqlite` |

Matched as a whole path segment or a suffix (e.g. `user_service` → service,
module `user`). Hexagonal structural dirs (`adapter`, `port`, `inbound`,
`outbound`, …) are treated as containers so the module resolves to the
bounded-context name. Classification is relative to your module path, so a
module name that happens to contain a keyword never mis-classifies. Extend via
`Options.Classify`.

## Auto-layer (naming-agnostic)

Keyword classification needs conventional package names. Real codebases —
especially microservices, each with its own conventions — often don't follow
them. So archview reads the **call chain** instead, **on by default**: starting
from each detected endpoint it walks the graph, includes every function actually
reached, and infers a layer from each one's role in the chain:

- the **entry** (the endpoint's handler) → controller,
- a function that **calls onward** into the app → service,
- a **sink** (one that only reaches external/leaf code) → repository.

So a service shows up because something *flows through it*, not because a folder
is named `service`. The flow is always readable — even with no layer
conventions at all, and the same analysis works across microservices that each
name things differently. Keyword classification still takes precedence where it
applies, so well-named projects are unchanged. Set `DisableAutoLayer: true` for
the curated keyword-only view.

## Frameworks

Route extraction is per-framework; an extractor only runs on a package that uses
its framework.

| Framework | Detected |
|-----------|----------|
| net/http  | `mux.HandleFunc("GET /path", h)` (Go 1.22 method patterns) |
| gin       | `r.GET/POST/...`, `Handle`, `Any` on `*gin.Engine` / `*gin.RouterGroup` (+ group prefixes) |
| echo      | `e.GET/POST/...`, `Any` on `*echo.Echo` / `*echo.Group` (+ group prefixes) |
| gRPC      | `Register<Svc>Server(reg, impl)` — each RPC method becomes an endpoint |
| GraphQL   | gqlgen `Query/Mutation/SubscriptionResolver` — each field becomes an endpoint |

gRPC and GraphQL detection is structural (the generated shapes), so it works
with real `google.golang.org/grpc` / gqlgen as-is. gin and echo join
`Group("/api")` prefixes onto routes. Add a framework by implementing the
`archview.Extractor` interface and passing it in `Options.Extractors`.

## Outbound ports (hexagonal)

With `ShowPorts: true`, an interface implemented by a repository-layer adapter
is surfaced as a **port** node — the seam between the core and an adapter:

```
service ──uses──▶ OrderRepository (port) ◀──implements── postgres adapter
```

Both arrows converge on the interface (dependency inversion). The direct
service→repository edge is replaced by the two edges through the port.

## Buses (CQRS / event-driven)

A command/query/event mediator stores handlers in a map keyed at runtime, so a
plain call graph can only over-approximate (every dispatch fans out to every
handler). With `DetectBuses: true`, archview reads the registration sites
(`bus.Register(Cmd{}.Name(), NewHandler(...))`, `Subscribe(...)`) and draws the
**precise** `caller → handler` routing instead — including event fan-out to the
exact subscribers. Marker methods and the bus internals drop out.

## UI

Swimlane per module, columns per layer. Drag a lane to move a whole module,
drag a node to nudge it, click a lane header to collapse/expand, pan/zoom,
hover to highlight a node's neighbors, and use the toolbar (Fit / Expand all /
Collapse all / zoom).

## Examples

```sh
# modular MVC over gin
go -C examples/gin-mvc run -buildvcs=false .      # http://localhost:8080/graph

# hexagonal / ports & adapters over net/http (ShowPorts)
go -C examples/hexagonal run -buildvcs=false .    # http://localhost:8090/graph

# CQRS + command/query/event buses (DetectBuses)
go -C examples/cqrs run -buildvcs=false .         # http://localhost:8095/graph

# gRPC in clean architecture
go -C examples/grpc run -buildvcs=false .         # http://localhost:8096/graph

# GraphQL (gqlgen-style resolvers)
go -C examples/graphql run -buildvcs=false .      # http://localhost:8097/graph

# echo + /api group (archview served on :9098)
go -C examples/echo run -buildvcs=false .         # http://localhost:9098/graph
```

## Limitations

- Concrete calls + single-implementation interfaces (CHA). Multiple
  implementations of one interface over-approximate the edges — unless a port
  or bus mediates them (`ShowPorts` / `DetectBuses`).
- gin/echo group prefixes are joined; net/http `StripPrefix` and chi `Route`
  nesting are not yet.
- net/http + gin + echo + gRPC + GraphQL; dev-live only (no baked artifact).
- Inline closure handlers create an endpoint node but no handler link.
- Unexported free-function helpers are collapsed through by default
  (`ShowHelpers` to keep them).

## Roadmap

layer-violation linter · outbound/external call detection (DB/HTTP/queue) ·
WebSocket extractor · fiber/chi adapters · `archview.yaml` config · prod-baked
(`go:generate` + `go:embed`) · JetBrains editor · search/filter. See
[`docs/prd.md`](docs/prd.md).
