import './global.css'
import 'virtual:fonts.css'
import 'react-toastify/dist/ReactToastify.css'

import React from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { HelmetProvider } from 'react-helmet-async'

import './i18n'

import Fullscreen from './layouts/Fullscreen'
import Navout from './layouts/Navout'
import App from './components/App'
import Login from './components/Login'
import Logout from './components/Logout'
import NotFound from './components/404'

createRoot(document.getElementById('root') as Element).render(
  <React.StrictMode>
    <HelmetProvider>
      <BrowserRouter>
        <Routes>
          <Route element={<Navout />}>
            <Route
              index
              element={<App />}
            />
          </Route>
          <Route element={<Fullscreen />}>
            <Route
              element={<Login />}
              path="/login"
            />
            <Route
              element={<Logout />}
              path="/logout"
            />
            <Route
              element={<NotFound />}
              path="*"
            />
          </Route>
        </Routes>
      </BrowserRouter>
    </HelmetProvider>
  </React.StrictMode>
)
