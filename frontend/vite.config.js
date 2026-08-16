import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { compression } from 'vite-plugin-compression2'

export default defineConfig({
  plugins: [
    vue(),
    compression({
      algorithm: 'gzip',
      include: [/\.(js|css|json|svg|txt)$/],
      threshold: 1024,
      deleteOriginalAssets: false
    }),
    compression({
      algorithm: 'brotliCompress',
      include: [/\.(js|css|json|svg|txt)$/],
      threshold: 1024,
      deleteOriginalAssets: false
    })
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
