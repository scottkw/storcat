import React from 'react'
import {createRoot} from 'react-dom/client'
import './style.css'
import './index.css'
import './workspace.css'
import { initThemeTokens } from './themeTokens'
import App from './App'

// Must run synchronously, before createRoot's render call -- a post-mount
// effect fires after first paint and reintroduces the launch flash.
initThemeTokens()

const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    <React.StrictMode>
        <App/>
    </React.StrictMode>
)
