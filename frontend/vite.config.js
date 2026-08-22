import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { compression } from 'vite-plugin-compression2'

const enableBrotli = process.env.SING_PANEL_BROTLI === '1'

export default defineConfig({
  plugins: [
    vue(),
    compression({
      algorithm: 'gzip',
      include: [/\.(js|css|json|svg|txt)$/],
      threshold: 1024,
      deleteOriginalAssets: false
    }),
    ...(enableBrotli ? [compression({
      algorithm: 'brotliCompress',
      include: [/\.(js|css|json|svg|txt)$/],
      threshold: 1024,
      deleteOriginalAssets: false
    })] : [])
  ],
  server: {
    host: '0.0.0.0',
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/clash_api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        ws: true
      }
    }
  },
  build: {
    outDir: 'dist',
    assetsDir: 'assets'
  }
})
