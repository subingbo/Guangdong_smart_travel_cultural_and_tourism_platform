import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      // Dev proxy: point at either backend by changing the target.
      '/api': {
        target: process.env.API_PROXY_TARGET || process.env.VITE_API_BASE_URL || 'http://localhost:8081',
        changeOrigin: true,
      },
      // Uploaded blog/avatar images are served by the Go backend.
      '/uploads': {
        target: process.env.API_PROXY_TARGET || process.env.VITE_API_BASE_URL || 'http://localhost:8081',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
  },
  preview: {
    port: 4173,
    proxy: {
      '/api': {
        target: process.env.API_PROXY_TARGET || process.env.VITE_API_BASE_URL || 'http://localhost:8081',
        changeOrigin: true,
      },
      '/uploads': {
        target: process.env.API_PROXY_TARGET || process.env.VITE_API_BASE_URL || 'http://localhost:8081',
        changeOrigin: true,
      },
    },
  },
});
