import fs from 'fs';
import path from 'path';

import { reactRouter } from '@react-router/dev/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';
import tsconfigPaths from 'vite-tsconfig-paths';

const PATH_TO_CLIENT_PACKAGE_JSON = path.join(process.cwd(), 'package.json');
const PATH_TO_SERVER_PACKAGE_JSON = path.join(process.cwd(), '../server/package.json');

const { version: clientVersion } = JSON.parse(fs.readFileSync(PATH_TO_CLIENT_PACKAGE_JSON, 'utf-8'));
const { version: serverVersion } = JSON.parse(fs.readFileSync(PATH_TO_SERVER_PACKAGE_JSON, 'utf-8'));

export default defineConfig({
  base: process.env.VITE_BASE_PATH || '/',
  plugins: [tailwindcss(), reactRouter(), tsconfigPaths()],
  define: {
    'import.meta.env.VERSION': process.env.NODE_ENV === 'development' ? undefined : JSON.stringify(`${clientVersion}-${serverVersion}`),
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
    },
  },
});
