import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  build: {
    // toolgui-wails/assets.go embeds this directory, and wails serves it from
    // its own origin. Keep the path in sync there and in
    // scripts/stub-assets.sh.
    outDir: '../../toolgui-wails/frontend/dist',
    emptyOutDir: true,
  },
})
