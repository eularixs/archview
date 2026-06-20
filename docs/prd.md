# PRD — archview

> Product Requirements Document · Versi 0.1 · 2026-06-20

## 1. Ringkasan

**archview** = library Go untuk visualisasi arsitektur & flow logic backend otomatis. Di-load di `main.go`, expose handler (`/graph` default). Buka di browser → flowchart auto-generate: module → controller → service → repository, lengkap panah call + endpoint. Framework-agnostic (net/http, gin) & pattern-aware (MVC, hexagonal).

## 2. Masalah

- Backend Go gede susah dibaca alur logic-nya tanpa baca file satu-satu.
- Onboarding lama, diagram manual cepat basi.
- Tool existing (`go-callvis`) cuma call graph mentah — ga sadar layer/endpoint, ga bisa klik-ke-source.

## 3. Goals

- G1 auto-generate tanpa anotasi · G2 layer-aware · G3 route-aware · G4 pattern-aware (panah ngikut pattern) · G5 framework-agnostic (adapter) · G6 click-to-source · G7 zero/low config.

## 4. Non-Goals (v0.1)

- Bukan profiler/tracer runtime (static analysis).
- Ga semua framework sekaligus (net/http + gin dulu).
- Ga ada auth bawaan di `/graph`.

## 5. Fitur inti

- Graph rendering `/graph` (swimlane module, kolom layer, drag, collapse, pan/zoom).
- Layer classification (heuristik + config).
- Route extraction (adapter per framework).
- Call graph (CHA).
- Click-to-source (vscode/cursor).
- Run modes: dev-live (v0.1), prod-baked (fase 2).

## 6. Pendekatan teknis

```
source → go/packages → go/callgraph (CHA) → classify → route adapter → graph.json → web /graph
```

Prior art: **go-callvis**. archview = go-callvis + route-aware + layer-aware + click-to-source + web UI modern.

## 7. Hard parts

| Risiko | Mitigasi |
|---|---|
| Interface dispatch (hexagonal port→adapter) | CHA (single-impl akurat); pointer analysis fase 2 |
| Route detection per-framework | Adapter terpisah; base net/http |
| Layer classification ambigu | Heuristik + config override |
| Source ga ada di runtime | dev-live; prod-baked fase 2 |

## 8. Roadmap

- v0.1 — dev-live, net/http+gin, MVC+hexagonal, vscode/cursor, 2 example.
- v0.2 — outbound detection, layer-violation linter, search/filter.
- v0.3 — echo/fiber/chi, archview.yaml config.
- v0.4 — prod-baked, JetBrains, pointer analysis.

## 9. Success criteria (v0.1)

- Example: `/graph` nampilin flow benar, klik node → editor loncat, setup ≤ 3 baris.
