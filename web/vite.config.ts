import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  build: {
    // dist/.gitkeep mantiene el directorio para go:embed; el script build
    // limpia dist/assets antes de buildear.
    emptyOutDir: false,
  },
  server: {
    proxy: { '/api': 'http://127.0.0.1:4321' },
  },
})
