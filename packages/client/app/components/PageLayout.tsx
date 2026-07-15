import type { ReactNode } from 'react';
import { useLocation } from 'react-router';

import { VersionHoverCard } from '~/components/VersionHoverCard';

export function PageLayout({ children }: { children: ReactNode }) {
  return (
    <div className="flex flex-col max-w-screen sm:max-w-[430px] w-full h-full">
      <div className="bg-dark-bg-primary text-dark-text-primary text-sm w-full">
        <nav className="flex p-2 border-b border-gray-700 justify-between items-center">
          <ul className="flex gap-4">
            <li>
              <PathAwareLink href="/">Play</PathAwareLink>
            </li>
            <li>
              <PathAwareLink href="/leaderboard">Leaderboard</PathAwareLink>
            </li>
            <li>
              <PathAwareLink href="/about">About</PathAwareLink>
            </li>
          </ul>

          <VersionHoverCard />
        </nav>
      </div>

      <main className="flex flex-col items-center justify-center p-2 bg-dark-bg-primary w-full flex-1">{children}</main>
    </div>
  );
}

function PathAwareLink({ href, children }: { href: string; children: string }) {
  const { pathname } = useLocation();
  const additionalClass = pathname === href ? 'font-bold!' : '';

  return (
    <a
      href={href}
      className={'navigation hover:text-dark-interactive-primary transition duration-200 ' + additionalClass}
      data-ga-value={children}
      data-ga-label="ga-navbar-link"
    >
      {children}
    </a>
  );
}
