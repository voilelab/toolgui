import React from 'react'
import ReactDOM from 'react-dom/client'
import Wails from '@wailsapp/runtime'

// The browser build pulls these off a CDN from index.html. A desktop app has
// to work offline, so they are bundled into the single CSS file instead.
import 'bulma/css/bulma.min.css'
// Font Awesome's base classes and animations only: solid.css carries the
// webfont, which a data: URL document will not load. icons.css draws the four
// icons the components use instead.
import '@fortawesome/fontawesome-free/css/fontawesome.min.css'
import './icons.css'
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
