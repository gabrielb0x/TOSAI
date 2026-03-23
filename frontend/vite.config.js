import { extname } from 'node:path'
import { defineConfig, loadEnv } from 'vite'

const truthyValues = new Set(['1', 'true', 'yes', 'on'])

const isEnabled = (value) => truthyValues.has(String(value || '').trim().toLowerCase())

const hashOnlyBuildOutput = {
  entryFileNames: 'assets/[hash].js',
  chunkFileNames: 'assets/[hash].js',
  assetFileNames: (assetInfo) => {
    const sourceName = assetInfo?.name || ''
    const extension = extname(sourceName)
    return `assets/[hash]${extension}`
  },
}

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  const backendTarget = env.VITE_BACKEND_PROXY_TARGET || 'http://localhost:9000'
  const maskBuildFilenames = isEnabled(env.VITE_MASK_BUILD_FILENAMES)
  const buildConfig = maskBuildFilenames
    ? {
        rollupOptions: {
          output: hashOnlyBuildOutput,
        },
      }
    : {}

  return {
    appType: 'spa',
    server: {
      host: true,
      port: 5173,
      strictPort: true,
      proxy: {
        '/summary': {
          target: backendTarget,
          changeOrigin: true,
        },
        '/v1': {
          target: backendTarget,
          changeOrigin: true,
        },
      },
    },
    preview: {
      host: true,
      port: 4173,
      strictPort: true,
    },
    build: buildConfig,
  }
})
