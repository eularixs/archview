import Link from 'next/link';
import type { ReactNode } from 'react';
import {
  Layers,
  Webhook,
  MousePointerClick,
  Hexagon,
  Boxes,
  Wand2,
  ArrowRight,
} from 'lucide-react';

function GithubIcon({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" className={className} aria-hidden>
      <path d="M12 .5a12 12 0 0 0-3.79 23.39c.6.11.82-.26.82-.58v-2c-3.34.73-4.04-1.61-4.04-1.61-.55-1.39-1.34-1.76-1.34-1.76-1.09-.75.08-.73.08-.73 1.2.09 1.84 1.24 1.84 1.24 1.07 1.83 2.81 1.3 3.5.99.11-.78.42-1.3.76-1.6-2.66-.3-5.47-1.33-5.47-5.93 0-1.31.47-2.38 1.24-3.22-.13-.3-.54-1.52.12-3.18 0 0 1.01-.32 3.3 1.23a11.5 11.5 0 0 1 6 0c2.29-1.55 3.3-1.23 3.3-1.23.66 1.66.25 2.88.12 3.18.77.84 1.24 1.91 1.24 3.22 0 4.61-2.81 5.62-5.49 5.92.43.37.81 1.1.81 2.22v3.29c0 .32.22.7.83.58A12 12 0 0 0 12 .5Z" />
    </svg>
  );
}

export default function HomePage() {
  return (
    <main className="flex flex-col">
      <Hero />
      <Features />
      <MountSection />
      <Pipeline />
      <Footer />
    </main>
  );
}

/* ----------------------------------------------------------------------- */

function Hero() {
  return (
    <section className="relative overflow-hidden border-b border-fd-border">
      <div className="av-dotgrid pointer-events-none absolute inset-0 opacity-70 [mask-image:radial-gradient(120%_90%_at_50%_0%,black,transparent_75%)]" />
      <div
        className="pointer-events-none absolute -top-40 left-1/2 h-[28rem] w-[44rem] -translate-x-1/2 rounded-full blur-3xl"
        style={{
          background:
            'radial-gradient(closest-side, color-mix(in oklab, var(--av-controller) 22%, transparent), transparent)',
        }}
      />

      <div className="relative mx-auto grid max-w-6xl items-center gap-12 px-6 py-20 lg:grid-cols-[1.05fr_1fr] lg:py-28">
        <div>
          <span
            className="av-rise inline-flex items-center gap-2 rounded-full border border-fd-border bg-fd-card/60 px-3 py-1 font-mono text-xs text-fd-muted-foreground backdrop-blur"
            style={{ animationDelay: '0ms' }}
          >
            <span
              className="h-1.5 w-1.5 rounded-full"
              style={{ background: 'var(--av-service)' }}
            />
            Go library · static analysis · v0.0.x beta
          </span>

          <h1
            className="av-rise mt-6 text-balance text-5xl font-extrabold leading-[1.05] sm:text-6xl"
            style={{ animationDelay: '70ms' }}
          >
            See your backend&apos;s{' '}
            <span style={{ color: 'var(--av-controller)' }}>flow</span>,
            <br />
            drawn from the code.
          </h1>

          <p
            className="av-rise mt-6 max-w-lg text-lg text-fd-muted-foreground"
            style={{ animationDelay: '140ms' }}
          >
            archview reads your Go source and renders the live path from{' '}
            <Layer c="endpoint">endpoint</Layer> →{' '}
            <Layer c="controller">controller</Layer> →{' '}
            <Layer c="service">service</Layer> →{' '}
            <Layer c="repository">repository</Layer>. No annotations. Click any
            node to open it in your editor.
          </p>

          <div
            className="av-rise mt-8 flex flex-wrap items-center gap-3"
            style={{ animationDelay: '210ms' }}
          >
            <Link
              href="/docs"
              className="group inline-flex items-center gap-2 rounded-lg bg-fd-primary px-5 py-2.5 text-sm font-semibold text-fd-primary-foreground transition-transform hover:-translate-y-0.5"
            >
              Get started
              <ArrowRight className="size-4 transition-transform group-hover:translate-x-0.5" />
            </Link>
            <a
              href="https://github.com/eularixs/archview"
              className="inline-flex items-center gap-2 rounded-lg border border-fd-border px-5 py-2.5 text-sm font-semibold transition-colors hover:bg-fd-accent"
            >
              <GithubIcon className="size-4" />
              GitHub
            </a>
          </div>

          <div
            className="av-rise mt-7 inline-flex items-center gap-3 rounded-lg border border-fd-border bg-fd-card/70 px-4 py-2.5 font-mono text-sm backdrop-blur"
            style={{ animationDelay: '280ms' }}
          >
            <span className="select-none text-fd-muted-foreground">$</span>
            <span>go get github.com/eularixs/archview</span>
          </div>
        </div>

        <div
          className="av-rise relative"
          style={{ animationDelay: '180ms' }}
        >
          <HeroGraph />
        </div>
      </div>
    </section>
  );
}

function Layer({
  c,
  children,
}: {
  c: 'endpoint' | 'controller' | 'service' | 'repository';
  children: ReactNode;
}) {
  return (
    <span className="font-medium" style={{ color: `var(--av-${c})` }}>
      {children}
    </span>
  );
}

/* The signature visual: an architecture graph drawn the way archview draws it. */
function HeroGraph() {
  const nodes = [
    { x: 6, y: 158, c: 'endpoint', label: 'GET /users', sub: 'endpoint' },
    { x: 150, y: 158, c: 'controller', label: 'UserController', sub: 'controller' },
    { x: 294, y: 158, c: 'service', label: 'UserService', sub: 'service' },
    { x: 438, y: 92, c: 'repository', label: 'postgres', sub: 'repository' },
    { x: 438, y: 224, c: 'repository', label: 'gateway', sub: 'repository' },
  ] as const;
  const W = 132;
  const H = 52;
  const edges = [
    [0, 1],
    [1, 2],
    [2, 3],
    [2, 4],
  ] as const;
  const cx = (i: number) => nodes[i].x;
  const path = (a: number, b: number) => {
    const x1 = cx(a) + W;
    const y1 = nodes[a].y + H / 2;
    const x2 = cx(b);
    const y2 = nodes[b].y + H / 2;
    const dx = Math.max(34, Math.abs(x2 - x1) * 0.5);
    return `M${x1},${y1} C${x1 + dx},${y1} ${x2 - dx},${y2} ${x2},${y2}`;
  };

  return (
    <svg
      viewBox="0 0 588 340"
      className="w-full drop-shadow-[0_24px_60px_rgba(0,0,0,0.45)]"
      role="img"
      aria-label="Architecture graph: endpoint to controller to service to repository"
    >
      <rect
        x="0"
        y="0"
        width="588"
        height="340"
        rx="16"
        className="fill-fd-card/40 stroke-fd-border"
        strokeWidth="1"
      />

      {/* edges */}
      {edges.map(([a, b], i) => {
        const color = `var(--av-${nodes[b].c})`;
        return (
          <g key={i}>
            <path
              d={path(a, b)}
              fill="none"
              stroke="currentColor"
              className="text-fd-border"
              strokeWidth="2"
            />
            <path
              d={path(a, b)}
              fill="none"
              stroke={color}
              strokeWidth="2"
              className="av-flow"
              style={{ animationDelay: `${i * 200}ms`, opacity: 0.9 }}
            />
          </g>
        );
      })}

      {/* nodes — outer <g> positions (attribute), inner <g> animates (CSS),
          so the rise animation never overrides the translate. */}
      {nodes.map((n, i) => {
        const color = `var(--av-${n.c})`;
        return (
          <g key={i} transform={`translate(${n.x},${n.y})`}>
            <g
              className="av-rise"
              style={{ animationDelay: `${300 + i * 110}ms` }}
            >
              <rect
                width={W}
                height={H}
                rx="10"
                className="fill-fd-background stroke-fd-border"
                strokeWidth="1"
              />
              <rect x="0" y="8" width="4" height={H - 16} rx="2" fill={color} />
              <circle cx={W - 13} cy="14" r="3" fill={color} />
              <text
                x="14"
                y="22"
                className="fill-fd-foreground"
                style={{ font: '600 12px var(--font-mono)' }}
              >
                {n.label}
              </text>
              <text
                x="14"
                y="38"
                style={{ font: '500 9px var(--font-mono)', fill: color }}
              >
                {n.sub}
              </text>
            </g>
          </g>
        );
      })}
    </svg>
  );
}

/* ----------------------------------------------------------------------- */

const features = [
  {
    icon: Layers,
    c: 'controller',
    title: 'Layer-aware',
    body: 'Each function placed as controller, service or repository — grouped into module swimlanes.',
  },
  {
    icon: Webhook,
    c: 'endpoint',
    title: 'Route-aware',
    body: 'Endpoints detected from your gin and net/http registrations, wired to their handlers.',
  },
  {
    icon: MousePointerClick,
    c: 'service',
    title: 'Click to source',
    body: 'Every node deep-links to file:line in VS Code or Cursor.',
  },
  {
    icon: Hexagon,
    c: 'repository',
    title: 'Pattern-aware',
    body: 'Modular MVC or hexagonal ports & adapters — arrows follow your real call graph.',
  },
  {
    icon: Boxes,
    c: 'controller',
    title: 'Framework-agnostic',
    body: 'gin and net/http out of the box; add any framework with a single interface.',
  },
  {
    icon: Wand2,
    c: 'endpoint',
    title: 'Zero annotations',
    body: 'Pure static analysis via a CHA call graph. No tags, no comments, no decorators.',
  },
] as const;

function Features() {
  return (
    <section className="mx-auto w-full max-w-6xl px-6 py-20">
      <p className="font-mono text-xs uppercase tracking-[0.2em] text-fd-muted-foreground">
        What you get
      </p>
      <h2 className="mt-3 max-w-2xl text-3xl font-bold sm:text-4xl">
        go-callvis, plus the three things it never knew.
      </h2>
      <div className="mt-12 grid gap-px overflow-hidden rounded-2xl border border-fd-border bg-fd-border sm:grid-cols-2 lg:grid-cols-3">
        {features.map((f) => {
          const Icon = f.icon;
          return (
            <div
              key={f.title}
              className="group relative bg-fd-background p-6 transition-colors hover:bg-fd-card"
            >
              <span
                className="inline-flex size-10 items-center justify-center rounded-lg border border-fd-border"
                style={{
                  color: `var(--av-${f.c})`,
                  background: `color-mix(in oklab, var(--av-${f.c}) 12%, transparent)`,
                }}
              >
                <Icon className="size-5" />
              </span>
              <h3 className="mt-4 text-lg font-semibold">{f.title}</h3>
              <p className="mt-1.5 text-sm text-fd-muted-foreground">
                {f.body}
              </p>
              <span
                className="absolute inset-x-0 bottom-0 h-0.5 scale-x-0 transition-transform duration-300 group-hover:scale-x-100"
                style={{ background: `var(--av-${f.c})`, transformOrigin: 'left' }}
              />
            </div>
          );
        })}
      </div>
    </section>
  );
}

/* ----------------------------------------------------------------------- */

function MountSection() {
  return (
    <section className="border-y border-fd-border bg-fd-card/30">
      <div className="mx-auto grid max-w-6xl items-center gap-10 px-6 py-20 lg:grid-cols-2">
        <div>
          <p className="font-mono text-xs uppercase tracking-[0.2em] text-fd-muted-foreground">
            Three lines
          </p>
          <h2 className="mt-3 text-3xl font-bold sm:text-4xl">
            Mount it. Open <span style={{ color: 'var(--av-controller)' }}>/graph</span>.
          </h2>
          <p className="mt-4 max-w-md text-fd-muted-foreground">
            archview owns its own path and hands the rest back to your router.
            It analyzes the module at startup, so the graph is always current
            with your code.
          </p>
          <Link
            href="/docs/quickstart"
            className="mt-6 inline-flex items-center gap-2 text-sm font-semibold"
            style={{ color: 'var(--color-fd-primary)' }}
          >
            Read the quickstart
            <ArrowRight className="size-4" />
          </Link>
        </div>

        <div className="overflow-hidden rounded-xl border border-fd-border bg-fd-background shadow-2xl">
          <div className="flex items-center gap-2 border-b border-fd-border px-4 py-2.5">
            <span className="size-3 rounded-full bg-fd-border" />
            <span className="size-3 rounded-full bg-fd-border" />
            <span className="size-3 rounded-full bg-fd-border" />
            <span className="ml-2 font-mono text-xs text-fd-muted-foreground">
              main.go
            </span>
          </div>
          <pre className="overflow-x-auto p-5 font-mono text-[13px] leading-relaxed">
            <code>
              <span className="text-fd-muted-foreground">{`// analyze source, serve the graph`}</span>
              {`\n`}
              <span style={{ color: 'var(--av-service)' }}>av</span>
              {`, _ := `}
              <span style={{ color: 'var(--av-controller)' }}>archview</span>
              {`.New(archview.Options{Root: `}
              <span style={{ color: 'var(--av-endpoint)' }}>{`"."`}</span>
              {`})\n`}
              <span style={{ color: 'var(--av-service)' }}>av</span>
              {`.Mount(mux)\n`}
              {`http.ListenAndServe(`}
              <span style={{ color: 'var(--av-endpoint)' }}>{`":8080"`}</span>
              {`, mux)`}
            </code>
          </pre>
        </div>
      </div>
    </section>
  );
}

/* ----------------------------------------------------------------------- */

function Pipeline() {
  const steps = ['source', 'CHA call graph', 'classify layers', 'extract routes', '/graph'];
  return (
    <section className="mx-auto w-full max-w-6xl px-6 py-20">
      <p className="font-mono text-xs uppercase tracking-[0.2em] text-fd-muted-foreground">
        The pipeline
      </p>
      <div className="mt-6 flex flex-wrap items-center gap-3">
        {steps.map((s, i) => (
          <div key={s} className="flex items-center gap-3">
            <span className="rounded-lg border border-fd-border bg-fd-card px-4 py-2 font-mono text-sm">
              {s}
            </span>
            {i < steps.length - 1 && (
              <ArrowRight className="size-4 text-fd-muted-foreground" />
            )}
          </div>
        ))}
      </div>
    </section>
  );
}

/* ----------------------------------------------------------------------- */

function Footer() {
  return (
    <footer className="border-t border-fd-border">
      <div className="mx-auto flex max-w-6xl flex-col items-center justify-between gap-4 px-6 py-10 text-sm text-fd-muted-foreground sm:flex-row">
        <span
          className="font-bold text-fd-foreground"
          style={{ fontFamily: 'var(--font-display), sans-serif' }}
        >
          archview
        </span>
        <nav className="flex items-center gap-6">
          <Link href="/docs" className="hover:text-fd-foreground">
            Docs
          </Link>
          <Link href="/docs/roadmap" className="hover:text-fd-foreground">
            Roadmap
          </Link>
          <a
            href="https://github.com/eularixs/archview"
            className="hover:text-fd-foreground"
          >
            GitHub
          </a>
        </nav>
        <span className="font-mono text-xs">© Eularix · MIT</span>
      </div>
    </footer>
  );
}
