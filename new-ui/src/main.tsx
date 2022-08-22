import './global.css'
import 'virtual:fonts.css'

import React from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Route, Routes } from 'react-router-dom'

import App from './components/App'

createRoot(document.getElementById('root') as Element).render(
  <React.StrictMode>
    <BrowserRouter>
      <Routes>
        <Route element={<App />} path="/" />
      </Routes>
    </BrowserRouter>
  </React.StrictMode>
)
