import React from 'react'
import ReactDOM from 'react-dom/client'

// After WailsApp: it pulls in the library's stylesheets — Bulma and Font
// Awesome included — and these override them.
import { WailsApp } from './WailsApp'
import './index.css'

// Wails injects its runtime and the bindings from the head of this page, so
// window.go and window.runtime are ready by the time this module runs.
const root = ReactDOM.createRoot(document.getElementById('root'))
root.render(
  <React.StrictMode>
    <WailsApp />
  </React.StrictMode>
)
