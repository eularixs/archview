import type { BaseLayoutProps } from 'fumadocs-ui/layouts/shared';
import { appName, gitConfig } from './shared';

function LogoMark() {
  return (
    <svg
      width="22"
      height="22"
      viewBox="0 0 22 22"
      fill="none"
      aria-hidden
    >
      <path
        d="M4 11h10.6M14.6 11l4.4-5M14.6 11l4.4 5"
        stroke="currentColor"
        strokeOpacity="0.32"
        strokeWidth="1.3"
      />
      <circle cx="4" cy="11" r="2.4" fill="var(--av-endpoint)" />
      <circle cx="9.3" cy="11" r="2.4" fill="var(--av-controller)" />
      <circle cx="14.6" cy="11" r="2.4" fill="var(--av-service)" />
      <circle cx="19" cy="6" r="2.1" fill="var(--av-repository)" />
      <circle cx="19" cy="16" r="2.1" fill="var(--av-repository)" />
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
