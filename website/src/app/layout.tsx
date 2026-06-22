import { RootProvider } from 'fumadocs-ui/provider/next';
import './global.css';
import type { Metadata } from 'next';
import {
  Funnel_Display,
  Hanken_Grotesk,
  JetBrains_Mono,
} from 'next/font/google';
import { appName, siteDescription, siteUrl } from '@/lib/shared';

export const metadata: Metadata = {
  metadataBase: new URL(siteUrl),
  title: {
    default: `${appName} — architecture flow graph for Go backends`,
    template: `%s · ${appName}`,
  },
  description: siteDescription,
  applicationName: appName,
  keywords: [
    'archview',
    'Go',
    'Golang',
    'architecture',
    'call graph',
    'static analysis',
    'hexagonal',
    'CQRS',
    'gin',
    'echo',
    'gRPC',
    'GraphQL',
    'visualization',
    'developer tools',
  ],
  authors: [{ name: 'Eularix' }],
  openGraph: {
    type: 'website',
    siteName: appName,
    url: siteUrl,
    title: `${appName} — architecture flow graph for Go backends`,
    description: siteDescription,
  },
  twitter: {
    card: 'summary_large_image',
    title: `${appName} — architecture flow graph for Go backends`,
    description: siteDescription,
  },
  robots: { index: true, follow: true },
};

const display = Funnel_Display({
  subsets: ['latin'],
  variable: '--font-display',
  display: 'swap',
});

const sans = Hanken_Grotesk({
  subsets: ['latin'],
  variable: '--font-sans',
  display: 'swap',
});

const mono = JetBrains_Mono({
  subsets: ['latin'],
  variable: '--font-mono',
  display: 'swap',
});

export default function Layout({ children }: LayoutProps<'/'>) {
  return (
    <html
      lang="en"
      className={`${display.variable} ${sans.variable} ${mono.variable}`}
      suppressHydrationWarning
    >
      <body className="flex flex-col min-h-screen">
        <RootProvider>{children}</RootProvider>
      </body>
    </html>
  );
}
