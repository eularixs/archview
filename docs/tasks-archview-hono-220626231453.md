# Tasks — archview-hono (TypeScript / Hono port)

> Dibuat: 2026-06-22 23:14 WIB · Status: PLAN/TASKS ONLY (belum dibangun)
> Repo: `eularixs/archview-hono` · Docs: tetap 1 di `eularixs/archview-docs`
> **Verifiable di sini** (bun + node ada).

## Prinsip
- Emit **graph.json schema yang SAMA** persis archview Go (kontrak lintas-bahasa).
- **Reuse UI** (`index.html`) tanpa ubah — serve + `/graph/data`.
- Analisa via **ts-morph** (wrap TypeScript compiler API: AST + type checker + symbol resolution = padanan Roslyn/go-packages).
- bun buat install/build/run/test. Tiap fitur verify lewat `/graph/data` + example.
- Identity Eularix, MIT, no Claude trailer.

## F0 — Scaffold + UI reuse + serve
- [ ] F0.1 `bun init` package `@eularix/archview` (atau `archview`), TS, repo init. tsconfig strict.
- [ ] F0.2 Meta: package.json (ESM), MIT, README/CONTRIBUTING/CoC/SECURITY (pola Go).
- [ ] F0.3 Graph model TS (`type Node/Edge/Graph`) — field identik schema Go.
- [ ] F0.4 Embed `index.html` (copy dari archview Go `web/index.html`) — bundle/inline.
- [ ] F0.5 Serve: Hono middleware `archview({ root, basePath })` → `GET {base}` HTML, `GET {base}/data` JSON. Exclude self-route. (Plus mode standalone `Bun.serve`.)
- [ ] F0.6 `Options`: root, basePath, editor, showPorts, detectBuses(→mediator), disableAutoLayer, lintLayers, showHelpers.
- [ ] F0.7 Smoke: serve graph dummy, buka `/graph`, UI reuse render.

## F1 — Core analysis (ts-morph)
- [ ] F1.1 Load project: `new Project({ tsConfigFilePath })` dari root, ambil source files (exclude node_modules/.d.ts).
- [ ] F1.2 Kumpulin functions/methods (`FunctionDeclaration`, `MethodDeclaration`, arrow assigned) + posisi `file:line` buat editorURL.
- [ ] F1.3 Call graph: per body, walk `CallExpression` → `getSymbol()`/`getResolvedSignature()` → callee declaration. Edge caller→callee (project-owned).
- [ ] F1.4 Interface dispatch (≈CHA): resolve call via interface ke impl (type checker `getImplementations` / cari class implements).
- [ ] F1.5 Classify layer: folder/naming (`controllers`/`handlers`/`routes`→controller, `services`/`usecases`→service, `repositories`/`db`/`prisma`/`drizzle`→repository) + module derive (feature-first & layer-first, port `moduleFor`).
- [ ] F1.6 **Auto-layer** (port BFS chain: entry=controller, calls-onward=service, sink=repository). Default ON.
- [ ] F1.7 Builder: nodes + call edges + collapse-helper + prune isolated + **prune-disconnected** + sort. Output graph.json.
- [ ] F1.8 Skip generated: file `// Code generated` / `.gen.ts` / Prisma client.

## F2 — Routes (Hono + lainnya)
- [ ] F2.1 **Hono**: `app.get/post/put/delete/patch/all('/path', handler)` + `app.route('/api', sub)` (prefix join, analog Group) + `new Hono()`. Handler = func/method ref → resolve declaration.
- [ ] F2.2 **Generic router** (reuse konsep Go): `router.<verb>(path, handler)` apa pun (Express `app.get`, Fastify, Elysia `.get`). Match by verb + string path + function arg.
- [ ] F2.3 Extractor interface TS + registry. Opt-out via Options.
- [ ] F2.4 Example `examples/hono-mvc/` + `examples/hexagonal/`. Verify endpoint→controller→service→repo.

## F3 — Depth: ports, mediator, RPC
- [ ] F3.1 **Outbound ports** (`ShowPorts`): interface di-implement repository-layer class → port node. Prisma/Drizzle repo adapter.
- [ ] F3.2 **Mediator/CQRS** (≈DetectBuses): deteksi handler registry (mis. custom bus, `tsyringe`/nestjs CQRS `@CommandHandler`, atau map-based) + dispatch → edge presisi. (Pola registrasi → routing.)
- [ ] F3.3 **tRPC / RPC**: `router({ procedure })` / `.query`/`.mutation` → tiap procedure = endpoint. (tRPC = "RPC" dunia TS.)
- [ ] F3.4 Example cqrs/trpc + verify.

## F4 — Reach
- [ ] F4.1 **WebSocket**: Hono `upgradeWebSocket` / `Bun.serve` ws handler → label "WS".
- [ ] F4.2 Framework lain: Express, Fastify, Elysia, NestJS controllers (`@Get()` decorator).
- [ ] F4.3 GraphQL (Apollo/Yoga resolvers) — opsional.

## F5 — Polish & parity
- [ ] F5.1 **Layer-violation linter** (`LintLayers`): reverse/skip/cross-module.
- [ ] F5.2 **Trivial-helper filter** + endpoint click-to-source (registration site) + dedup multi-transport.
- [ ] F5.3 Click-to-source: `vscode://file/...` + `cursor://`.
- [ ] F5.4 CI (GitHub Actions): bun install + tsc + test + lint. Identity Eularix.
- [ ] F5.5 npm publish (manual approve).

## D — Docs unification (archview-docs)
- [ ] D1 Nav language switch / grup (`Go`, `TypeScript`, `.NET`).
- [ ] D2 Shared pages (semua): Intro, How it works, Auto-layer, Ports, Mediators, Linting, UI guide, **graph schema spec**.
- [ ] D3 Split per-bahasa: Frameworks (Hono/Express/Fastify/Elysia/NestJS/tRPC), Usage/Install, Examples. Tambah `/docs/ts/*`.

## Parity checklist vs Go (target 100%)
- [ ] Frameworks: generic router (Hono/Express/Fastify/Elysia) · tRPC (RPC) · WebSocket · NestJS decorator
- [ ] Auto-layer (chain, naming-agnostic, default ON) · ports · mediator/CQRS · linter
- [ ] Quality: helper filter · prefix join · generated skip · layer-first module · dedup · prune-disconnected · endpoint click-source · layout persistence (otomatis ikut UI reuse)

## Open decisions
- [ ] Nama npm: `@eularix/archview` vs `archview`.
- [ ] UI sharing: copy `index.html` (default) vs shared asset.
- [ ] ts-morph vs raw `typescript` API (saran: ts-morph, lebih enak).
- [ ] Target: Node 20+ / Bun. ESM-only?

## Urutan
F0 (scaffold+UI+serve dummy) → F1 (ts-morph core + auto-layer, INTI) → F2 (Hono + generic router) → F3 (ports + mediator + tRPC) → F4 (ws + framework lain) → F5 (linter + CI) → D (docs).
