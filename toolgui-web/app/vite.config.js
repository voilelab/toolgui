import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  build: {
    // ../web.go embeds app/build/index.html + app/build/static/*, and
    // tgexec serves them at / and /static/. Keep the layout in sync there,
    // in scripts/stub-assets.sh, and here.
    outDir: 'build',
    assetsDir: 'static',
  },
  server: {
    port: 3000,
  },
  test: {
    environment: 'jsdom',
    setupFiles: './src/setupTests.js',
  },
})
