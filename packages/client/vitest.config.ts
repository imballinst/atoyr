import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    environment: 'jsdom',
    include: ['app/**/*.test.{ts,tsx,js}'],
    setupFiles: ['app/test/setup.ts'],
  },
});
