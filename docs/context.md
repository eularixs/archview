# Context — archview

> Snapshot latar belakang + keputusan. Update: 2026-06-20

## Ide awal (dari Dimas)

Library Go general buat backend — ga peduli pakai framework atau nggak (gin/echo/fiber ke-cover). Saat di-load di `main.go`, render web di `/graph`: flowchart auto-generate dari logic. Modular MVC → grouping per module, di dalamnya controller/service/repository. Nampilin endpoint → controller func → (panah) service func → (panah) repo func. Panah otomatis ngikut pattern (hexagonal juga ikut, outbound kebaca). Tambahan: klik node → buka source di editor (VSCode/Cursor).

## Verdict feasibility

Possible. Core proven oleh `go-callvis` (callgraph + serve). archview = go-callvis + route-aware + layer-aware + click-to-source.

## Pipeline teknis

```
source → go/packages (AST+types) → callgraph (CHA) → classify layer
       → extract routes (adapter per framework) → emit graph.json → serve /graph
```

## Keputusan yang udah diambil

| Topik | Keputusan |
|---|---|
| Nama | **archview** (folder masih `gostruct/`) |
| Module path | `github.com/eularix/archview` (Eularix identity) |
| Run mode v0.1 | dev-live |
| Framework v0.1 | net/http + gin |
| Pattern v0.1 | modular MVC + hexagonal |
| Editor jump | vscode + cursor |
| Call graph algo | CHA (resolve single-impl interface) |

## Hard parts

| Masalah | Solusi |
|---|---|
| Interface dispatch (port→adapter hexagonal) | CHA (single-impl akurat); pointer analysis fase 2 |
| Route detection | Adapter per framework |
| Layer classification | Heuristik default + config |
| Source di runtime | dev-live; prod-baked fase 2 |

## Prior art

- `go-callvis` — call graph viz, foundation.
- `golang.org/x/tools/go/{packages,callgraph,ssa}` — toolchain inti.
- VSCode/Cursor URL scheme: `vscode://file/<path>:<line>:<col>`.

## Catatan insiden

2026-06-20 ~20:01: folder `Eularix/open-source/` ke-hapus dari disk di luar action Claude. Project di-regenerate penuh dari context percakapan (semua file ditulis sesi yang sama). Web UI juga selamet di `/tmp/page.html`.
