import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  // Library builds leave process.env.NODE_ENV to the consumer, but there is no
  // consumer here: the bundle is injected straight into a webview, where React
  // would blow up on the missing global.
  define: {
    'process.env.NODE_ENV': JSON.stringify('production'),
  },
  build: {
    // Wails v1 embeds exactly one JS and one CSS file (see
    // toolgui-wails/assets.go), so everything - fonts included - has to end up
    // in those two. Keep this in sync with scripts/stub-assets.sh.
    outDir: '../../toolgui-wails/assets',
    emptyOutDir: true,
    cssCodeSplit: false,
    assetsInlineLimit: Number.MAX_SAFE_INTEGER,
    lib: {
      entry: 'src/index.tsx',
      name: 'ToolGUIWails',
      formats: ['iife'],
      fileName: () => 'app.js',
      cssFileName: 'app',
    },
  },
})
