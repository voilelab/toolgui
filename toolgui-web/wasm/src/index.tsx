import React from 'react'
import ReactDOM from 'react-dom/client'

// After WasmApp: it pulls in the library's stylesheets, and these override
// them.
import { WasmApp } from './WasmApp'
import './index.css'
import { setupOffline } from './offline'

setupOffline().catch((e) => console.error('service worker', e))

const root = ReactDOM.createRoot(document.getElementById('root'))
root.render(
  <React.StrictMode>
    <WasmApp />
  </React.StrictMode>
)
