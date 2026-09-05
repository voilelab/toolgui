import React from 'react'
import ReactDOM from 'react-dom/client'
import Wails from '@wailsapp/runtime'

// The browser build pulls these off a CDN from index.html. A desktop app has
// to work offline, so they are bundled into the single CSS file instead.
import 'bulma/css/bulma.min.css'
// Only the solid set: the components use fa-moon, fa-sun, fa-spinner and
// fa-upload, and every unused webfont would be inlined into the CSS.
import '@fortawesome/fontawesome-free/css/fontawesome.min.css'
import '@fortawesome/fontawesome-free/css/solid.min.css'
import './index.css'

import { WailsApp } from './WailsApp'

// Wails injects its runtime into the webview and calls back once window.backend
// is bound. Its default HTML mounts the app at #app.
Wails.Init(() => {
  const root = ReactDOM.createRoot(document.getElementById('app'))
  root.render(
    <React.StrictMode>
      <WailsApp />
    </React.StrictMode>
  )
})
