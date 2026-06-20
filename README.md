# archview

Live architecture flow graph for Go backends. Mount it in `main.go`, open
`/graph` in a browser, and see your endpoints flow through
controller → service → repository — generated automatically from your source,
no annotations. Click any node to jump to its definition in your editor.

Framework-agnostic (net/http, gin) and pattern-aware (modular MVC, hexagonal):
the arrows follow the call graph of whatever layout you actually use.

> Status: **v0.1** — dev-live mode, net/http + gin, modular MVC + hexagonal.
> See [`docs/`](docs/) for the PRD, plan and roadmap.

## Install

```sh
go get github.com/eularix/archview
```

## Usage

```go
mux := http.NewServeMux()

av, err := archview.New(archview.Options{
    Root:     ".",        // module dir to analyze (default ".")
    BasePath: "/graph",   // mount path (default "/graph")
    Editor:   "vscode",   // click-to-source: "vscode" or "cursor"
})
if err != nil {
    log.Fatal(err)
}
av.Mount(mux)             // serves /graph and /graph/data

http.ListenAndServe(":8080", mux)
```

Open <http://localhost:8080/graph>.

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
bounded-context name. Extend via `Options.Classify`.

## UI

Swimlane per module, columns per layer. Drag a lane to move a whole module,
drag a node to nudge it, click a lane header to collapse/expand, pan/zoom,
hover to highlight a node's neighbors, and use the toolbar (Fit / Expand all /
Collapse all / zoom).

## Examples

```sh
# modular MVC over gin
go -C examples/gin-mvc run -buildvcs=false .      # http://localhost:8080/graph

# hexagonal / ports & adapters over net/http
go -C examples/hexagonal run -buildvcs=false .    # http://localhost:8090/graph
```

## Limitations (v0.1)

- Concrete calls + single-implementation interfaces (CHA). Multiple
  implementations of one interface over-approximate the edges.
- Route paths are the literal registered path; group/router prefixes are not
  joined yet (`/users`, not `/api/users`).
- net/http + gin only. dev-live only (no baked artifact).
- Inline closure handlers create an endpoint node but no handler link.

## Roadmap

outbound/external call detection (DB/HTTP/queue) · layer-violation linter ·
echo/fiber/chi adapters · `archview.yaml` config · prod-baked
(`go:generate` + `go:embed`) · JetBrains editor · search/filter. See
[`docs/prd.md`](docs/prd.md).
