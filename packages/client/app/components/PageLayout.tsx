import type { ReactNode } from 'react';
import { useLocation } from 'react-router';

export function PageLayout({ children }: { children: ReactNode }) {
  return (
    <div className="flex flex-col max-w-screen sm:max-w-[430px] w-full h-full">
      <div className="bg-dark-bg-primary text-dark-text-primary text-sm w-full">
        <nav className="mx-auto px-4 py-2 border-b border-gray-700">
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
        </nav>
      </div>

      <main className="flex flex-col items-center justify-center p-4 bg-dark-bg-primary w-full flex-1">{children}</main>
    </div>
  );
}

function PathAwareLink({ href, children }: { href: string; children: ReactNode }) {
  const { pathname } = useLocation();
  const additionalClass = pathname === href ? 'font-bold!' : '';

  return (
    <a href={href} className={'navigation ' + additionalClass} data-text={children}>
      {children}
    </a>
  );
}
