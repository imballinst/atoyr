import { InfoIcon } from 'lucide-react';
import { Popover } from 'radix-ui';
import type { ReactNode } from 'react';
import { useLocation } from 'react-router';

export function PageLayout({ children }: { children: ReactNode }) {
  return (
    <div className="flex flex-col max-w-screen sm:max-w-[430px] w-full h-full">
      <div className="bg-dark-bg-primary text-dark-text-primary text-sm w-full">
        <nav className="flex px-4 py-2 border-b border-gray-700 justify-between items-center">
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

          <Popover.Root>
            <Popover.Trigger asChild>
              <InfoIcon aria-label="Version" size={16} />
            </Popover.Trigger>
            <Popover.Portal>
              <Popover.Content
                className="bg-dark-bg-primary border border-dark-border-light px-1.5 py-1 rounded text-sm text-dark-text-primary outline-0 shadow"
                sideOffset={5}
                side="bottom"
                align="end"
              >
                {import.meta.env.DEV
                  ? 'Dev'
                  : (() => {
                      const [client, server] = import.meta.env.VERSION.split('-');
                      return (
                        <ul>
                          <li>Client version: {client}</li>
                          <li>Server version: {server}</li>
                        </ul>
                      );
                    })()}
              </Popover.Content>
            </Popover.Portal>
          </Popover.Root>
        </nav>
      </div>

      <main className="flex flex-col items-center justify-center p-4 bg-dark-bg-primary w-full flex-1">{children}</main>
    </div>
  );
}

function PathAwareLink({ href, children }: { href: string; children: string }) {
  const { pathname } = useLocation();
  const additionalClass = pathname === href ? 'font-bold!' : '';

  return (
    <a href={href} className={'navigation ' + additionalClass} data-ga-value={children} data-ga-label="ga-navbar-link">
      {children}
    </a>
  );
}
