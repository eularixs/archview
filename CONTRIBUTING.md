# Contributing to archview

Thanks for your interest in improving archview. This guide covers the library
repository; the documentation site lives in
[`eularixs/archview-docs`](https://github.com/eularixs/archview-docs).

## Development

```sh
go build ./...
go vet ./...
go test ./...
gofmt -l .            # must print nothing
```

Run an example end to end:

```sh
go -C examples/gin-mvc run -buildvcs=false .   # then open /graph
```

## Project layout

| Path         | Responsibility |
|--------------|----------------|
| `archview.go`| public API (`New`, `Options`, `Mount`, `Handler`) |
| `analyzer/`  | `go/packages` + SSA + CHA call graph, ports, bus detection |
| `route/`     | per-framework `Extractor`s (net/http, gin, echo, gRPC, GraphQL) |
| `classify/`  | layer + module classification |
| `build/`     | graph assembly |
| `web/`       | embedded `/graph` UI |
| `examples/`  | runnable demos (one Go module each) |

## Adding a framework

Implement the `Extractor` interface and add it to `route.Default()`:

```go
type Extractor interface {
    Name() string
    Match(pkg *packages.Package) bool
    Extract(pkg *packages.Package) []Route
}
```

Add a matching example under `examples/` and verify it via `/graph/data`.

## Pull requests

- Keep changes focused; one feature or fix per PR.
- New analysis features should be opt-in via `Options` and ship with an example.
- Run `gofmt`, `go vet`, `go test ./...`, and build every example before
  opening a PR — CI runs the same checks.
- Be honest about limitations in code comments and docs.

## Reporting issues

Open a GitHub issue with a minimal reproducer (a small module layout is ideal)
and the `/graph/data` output if relevant.

By contributing you agree your contributions are licensed under the
[MIT License](LICENSE).
