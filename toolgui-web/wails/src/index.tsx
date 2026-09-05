import React from 'react'
import ReactDOM from 'react-dom/client'

// The browser build pulls these off a CDN from index.html. A desktop app has
// to work offline, so they are bundled instead.
//
// bulma is pinned to the version app/index.html loads: the two have to render
// the same, and 1.0.4 turned the selected navbar item's text dark.
import 'bulma/css/bulma.min.css'
import '@fortawesome/fontawesome-free/css/fontawesome.min.css'
import '@fortawesome/fontawesome-free/css/solid.min.css'
import './index.css'

import { WailsApp } from './WailsApp'

// Wails injects its runtime and the bindings from the head of this page, so
// window.go and window.runtime are ready by the time this module runs.
const root = ReactDOM.createRoot(document.getElementById('root'))
root.render(
  <React.StrictMode>
    <WailsApp />
  </React.StrictMode>
)
