import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',
    port: Number(process.env.FRONT_PORT ?? 8081),
    strictPort: true,
    proxy: {
      '/api': {
        target: process.env.VITE_BACKEND_URL ?? 'http://127.0.0.1:8080',
        changeOrigin: true,
      },
    },
  },
});
