# CLAUDE.md — archview

Instruksi project-level. Tunduk ke global PAKEM di `/Users/dimasmaulana/CLAUDE.md` (literal reading, ask-don't-assume, no over-engineering, reuse-before-create, proof-before-done, no auto-commit/push). Yang di bawah = tambahan spesifik archview.

## Tentang repo

archview = library Go, visualisasi arsitektur/flow backend otomatis. Lihat `docs/prd.md` + `docs/context.md` + `docs/plan-*.md`.

## Aturan spesifik

1. **Static analysis, bukan runtime.** Core = `go/packages` + `go/callgraph` + `go/token`.
2. **Framework adapter = interface.** Tiap framework = satu impl `RouteExtractor`. Jangan if-else per framework di analyzer.
3. **Layer classification = heuristik + config.** Jangan hardcode nama module/layer project tertentu.
4. **Scope per fase.** Diminta A → kerjain A.
5. **Proof pakai example app.** Jangan klaim "done" tanpa verifikasi `/graph`.
6. **Editor URL.** `vscode://file/<abs>:<line>:<col>` + `cursor://`. Path absolut.
7. **Web assets embed.** `go:embed`. Self-contained.
8. **Public API minimal.** `New(opts)`, `Handler()`, `Mount(mux)`.
9. **Reuse x/tools.** `golang.org/x/tools/go/{packages,callgraph,ssa}`.
10. **Bahasa.** Caveman/terse ke Dimas. Docs & comment = normal.

## Layout package

```
archview.go        # public API
analyzer/          # AST + call graph
route/             # RouteExtractor + impl per framework
classify/          # layer + module heuristik
build/             # graph assembly
web/               # handler /graph + embed assets
examples/          # proof apps (gin-mvc, hexagonal)
docs/              # prd, context, plan, tasks
```

## Git

- Module path `github.com/eularix/archview` → Eularix identity (`Eularix <dimas.eularix@gmail.com>`, SSH `github.com-eularix`).
- JANGAN auto commit/push. JANGAN trailer `Co-Authored-By: Claude`.
