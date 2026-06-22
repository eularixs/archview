export const appName = 'archview';
export const version = 'v0.2.0';
export const docsRoute = '/docs';
export const docsImageRoute = '/og/docs';
export const docsContentRoute = '/llms.mdx/docs';

export const siteUrl =
  process.env.NEXT_PUBLIC_SITE_URL ?? 'https://archview-docs.vercel.app';

export const siteDescription =
  'archview renders a live architecture flow graph of a Go backend from source — endpoints flowing through controller → service → repository. Framework-agnostic (net/http, gin, echo, gRPC, GraphQL), pattern-aware (MVC, hexagonal, CQRS).';

export const gitConfig = {
  user: 'eularixs',
  repo: 'archview',
  branch: 'main',
};
