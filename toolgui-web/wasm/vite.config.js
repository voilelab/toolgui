import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  // Relative asset URLs, so the same build works at a site root and under a
  // project path like /toolgui/.
  base: './',
  build: {
    // scripts/build-wasm.sh copies this next to app.wasm and wasm_exec.js.
    outDir: 'build',
    assetsDir: 'static',
  },
})
