import { RootProvider } from 'fumadocs-ui/provider/next';
import './global.css';
import {
  Funnel_Display,
  Hanken_Grotesk,
  JetBrains_Mono,
} from 'next/font/google';

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
