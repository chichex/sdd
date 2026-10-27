import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
  build: {
    // dist/.gitkeep mantiene el directorio para go:embed; el script build
    // limpia dist/assets antes de buildear.
    emptyOutDir: false,
  },
  server: {
    proxy: { '/api': 'http://127.0.0.1:4321' },
  },
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    // El test de tokens lee index.css con ?raw: sin esto Vitest lo deja vacío.
    css: { include: [/index\.css/] },
  },
})
