import '@testing-library/jest-dom/vitest';
import { vi } from 'vitest';

// Source - https://stackoverflow.com/a/53449595
// Posted by Selrond, modified by community. See post 'Timeline' for change history
// Retrieved 2026-09-04, License - CC BY-SA 4.0
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation((query) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(), // Deprecated
    removeListener: vi.fn(), // Deprecated
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});
