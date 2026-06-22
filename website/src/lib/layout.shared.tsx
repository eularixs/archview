import type { BaseLayoutProps } from 'fumadocs-ui/layouts/shared';
import { appName, gitConfig } from './shared';

function LogoMark() {
  return (
    <svg width="24" height="24" viewBox="0 0 48 48" fill="none" aria-hidden>
      <rect width="48" height="48" rx="11" fill="#0d1320" />
      <g transform="translate(24 24) scale(0.92) translate(-24 -24)">
        <path
          d="M24 9 L10.5 39.5 M24 9 L37.5 39.5 M15.9 27.4 L32.1 27.4"
          stroke="#5B9BFF"
          strokeWidth="3.4"
          strokeLinecap="round"
          strokeLinejoin="round"
          fill="none"
        />
        <circle cx="24" cy="9" r="4" fill="#3B82F6" />
        <circle cx="10.5" cy="39.5" r="4" fill="#22D3EE" />
        <circle cx="37.5" cy="39.5" r="4" fill="#A855F7" />
      </g>
    </svg>
  );
}

export function baseOptions(): BaseLayoutProps {
  return {
    nav: {
      title: (
        <span
          className="inline-flex items-center gap-2 text-[15px] font-bold tracking-tight"
          style={{ fontFamily: 'var(--font-display), sans-serif' }}
        >
          <LogoMark />
          {appName}
        </span>
      ),
    },
    githubUrl: `https://github.com/${gitConfig.user}/${gitConfig.repo}`,
  };
}
