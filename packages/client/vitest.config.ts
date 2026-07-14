import tsconfigPaths from 'vite-tsconfig-paths';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  plugins: [tsconfigPaths()],
  test: {
    environment: 'jsdom',
    include: ['app/**/*.test.{ts,tsx,js}'],
    setupFiles: ['app/test/setup.ts'],
  },
});
