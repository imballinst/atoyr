import { Popover } from 'radix-ui';
import { useState } from 'react';
import { Link, useRouteLoaderData } from 'react-router';

import { computeCombinedVersion, getClientVersion, getServerVersion, writeStoredVersionInfo } from '~/lib/version';

const VERSION = import.meta.env.VERSION;
const COMBINED_VERSION = computeCombinedVersion(VERSION);

export function VersionPopover() {
  const rootData = useRouteLoaderData('root') as { showBadge: boolean } | undefined;
  const [badgeVisible, setBadgeVisible] = useState(rootData?.showBadge ?? false);

  function handleOpenChange(open: boolean) {
    if (open && badgeVisible) {
      setBadgeVisible(false);
      writeStoredVersionInfo(COMBINED_VERSION, new Date().toISOString());
    }
  }

  return (
    <Popover.Root onOpenChange={handleOpenChange}>
      <Popover.Trigger asChild>
        <button
          type="button"
          className="relative text-xs font-mono px-1.5 py-0.5 border border-dark-border-light rounded text-dark-text-secondary hover:text-dark-interactive-primary hover:border-dark-interactive-primary transition duration-200"
          aria-label={`Version ${COMBINED_VERSION}`}
        >
          {COMBINED_VERSION}
          {badgeVisible && <NewBadge />}
        </button>
      </Popover.Trigger>
      <Popover.Portal>
        <Popover.Content
          className="bg-dark-bg-primary border border-dark-border-light px-3 py-2 rounded text-sm text-dark-text-primary outline-0 shadow flex flex-col gap-1"
          sideOffset={5}
          side="bottom"
          align="end"
        >
          <span className="text-xs text-dark-text-secondary">Client version: {getClientVersion(VERSION)}</span>
          <span className="text-xs text-dark-text-secondary">Server version: {getServerVersion(VERSION)}</span>
          <Popover.Close asChild>
            <Link
              to="/changelog"
              className="text-xs text-dark-interactive-primary underline decoration-dotted hover:text-dark-interactive-primary/80 mt-0.5"
            >
              See what changed
            </Link>
          </Popover.Close>
        </Popover.Content>
      </Popover.Portal>
    </Popover.Root>
  );
}

function NewBadge() {
  return <span className="absolute -top-1.5 -right-1.5 bg-red-500 text-white text-[10px] font-bold px-1 rounded-sm leading-none">New</span>;
}
