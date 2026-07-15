import { XIcon } from 'lucide-react';
import { Dialog } from 'radix-ui';
import type { ReactNode } from 'react';

interface SharedDialogProps {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  title: string;
  trigger: ReactNode;
  children: ReactNode;
}

export function SharedDialog({ open, onOpenChange, title, trigger, children }: SharedDialogProps) {
  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Trigger asChild>{trigger}</Dialog.Trigger>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 bg-black/60" data-testid="dialog-overlay" />
        <Dialog.Content className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-dark-bg-secondary text-dark-text-primary rounded-2xl shadow-lg p-6 w-[calc(100%-2rem)] max-w-md max-h-[90vh] overflow-y-auto">
          <div className="flex items-center justify-between gap-4 mb-4">
            <Dialog.Title className="text-lg font-semibold">{title}</Dialog.Title>
            <Dialog.Close asChild>
              <button
                type="button"
                aria-label="Close"
                className="shrink-0 p-1 rounded text-dark-text-tertiary hover:text-dark-text-primary hover:bg-dark-bg-tertiary transition duration-200"
              >
                <XIcon size={20} />
              </button>
            </Dialog.Close>
          </div>
          {children}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
