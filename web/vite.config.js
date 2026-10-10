import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue()],
  base: './',
  resolve: { alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) } },
  server: { proxy: { '/api': 'http://127.0.0.1:8095', '/uploads': 'http://127.0.0.1:8095' } },
  build: {
    chunkSizeWarningLimit: 1500,
    rollupOptions: {
      output: {
        manualChunks: { 'element-plus': ['element-plus'], echarts: ['echarts'], xlsx: ['xlsx'] }
      }
    }
  }
})
