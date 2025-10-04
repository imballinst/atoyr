import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindPlugin from '@tailwindcss/vite';

export default defineConfig({
  plugins: [
    react(),
    // @ts-ignore
    tailwindPlugin(),
  ],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:4000',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
  },
});
