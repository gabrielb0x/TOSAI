import { defineConfig } from 'vite'

const backendTarget = process.env.VITE_BACKEND_PROXY_TARGET || 'http://localhost:9000'

export default defineConfig({
  server: {
    host: true,
    port: 5173,
    strictPort: true,
    proxy: {
      '/api': {
        target: backendTarget,
        changeOrigin: true,
      },
      '/healthz': {
        target: backendTarget,
        changeOrigin: true,
      },
    },
  },
})
