import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'
import viteCompression from 'vite-plugin-compression'

export default defineConfig({
  plugins: [
    vue(),
    viteCompression({
      algorithm: 'gzip',
      ext: '.gz',
      threshold: 10240
    })
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src')
    }
  },
  server: {
    port: 5173,
    host: '0.0.0.0',
    proxy: {
      '/api': 'http://localhost:8080'
    }
  },
  base: '/',
  build: {
    outDir: '../public',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) {
            return
          }

          if (id.includes('element-plus')) {
            return 'element-plus'
          }

          if (id.includes('axios')) {
            return 'http-vendor'
          }

          if (
            id.includes('vue') ||
            id.includes('pinia') ||
            id.includes('vue-router')
          ) {
            return 'vue-vendor'
          }

          return 'vendor'
        }
      }
    }
  }
}) 
