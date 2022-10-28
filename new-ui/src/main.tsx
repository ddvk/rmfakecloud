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
import Oops from './components/Oops'
import ResetPassword from './components/ResetPassword'
import Devices from './components/Devices'
import DocumentViewer from './components/DocumentViewer'
import Users from './components/Users'
import NewUser from './components/NewUser'
import EditUser from './components/EditUser'

createRoot(document.getElementById('root') as Element).render(
  <React.StrictMode>
    <HelmetProvider>
      <BrowserRouter>
        <Routes>
          <Route
            element={<Navout />}
            path="/"
          >
            <Route
              index
              element={<App />}
            />
            <Route
              element={<ResetPassword />}
              path="profile/reset_password"
            />
            <Route
              element={<Devices />}
              path="devices"
            />
            <Route
              element={<Users />}
              path="users"
            />
            <Route
              element={<NewUser />}
              path="users/new"
            />
            <Route
              element={<EditUser />}
              path="users/:userId/edit"
            />
          </Route>
          <Route element={<Fullscreen />}>
            <Route
              element={<DocumentViewer />}
              path="/documents/:docId/viewer"
            />
            <Route
              element={<Login />}
              path="/login"
            />
            <Route
              element={<Logout />}
              path="/logout"
            />
            <Route
              element={<Oops />}
              path="/oops"
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
