import './global.css'
import 'virtual:fonts.css'
import 'react-toastify/dist/ReactToastify.css'

import React from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Route, Routes } from 'react-router-dom'

import './i18n'

import Fullscreen from './layouts/Fullscreen'
import App from './components/App'
import Login from './components/Login'

createRoot(document.getElementById('root') as Element).render(
  <React.StrictMode>
    <BrowserRouter>
      <Routes>
        <Route element={<Fullscreen />}>
          <Route
            index
            element={<App />}
          />
          <Route
            element={<Login />}
            path="/login"
          />
        </Route>
      </Routes>
    </BrowserRouter>
  </React.StrictMode>
)
