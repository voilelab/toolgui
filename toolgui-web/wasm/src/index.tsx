import React from 'react'
import ReactDOM from 'react-dom/client'

// After WasmApp: it pulls in the library's stylesheets, and these override
// them.
import { WasmApp } from './WasmApp'
import './index.css'

const root = ReactDOM.createRoot(document.getElementById('root'))
root.render(
  <React.StrictMode>
    <WasmApp />
  </React.StrictMode>
)
