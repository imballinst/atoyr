import tsconfigPaths from 'vite-tsconfig-paths';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  plugins: [tsconfigPaths()],
  test: {
    environment: 'jsdom',
    // Node.js 26 ships an experimental Web Storage global whose localStorage
    // getter returns undefined unless --localstorage-file is set. It shadows
    // the localStorage that jsdom installs, so disable it for tests.
    execArgv: ['--no-experimental-webstorage'],
    include: ['app/**/*.test.{ts,tsx,js}'],
    setupFiles: ['app/test/setup.ts'],
  },
});
