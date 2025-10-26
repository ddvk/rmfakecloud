import {useEffect} from "react";
import { BrowserRouter as Router, Route, Switch } from "react-router-dom";
import { ToastContainer } from 'react-toastify';

import apiService from "./services/api.service";
import { AuthProvider } from "./common/useAuthContext";
import Role from "./common/Role";
import { PrivateRoute } from "./components/PrivateRoute";
import Navigationbar from "./components/Navigation";

import Login from "./pages/Login";
import Home from "./pages/Home";
import Connect from "./pages/Connect";
import Documents from "./pages/Documents";
import Integrations from "./pages/Integrations";
import Profile from "./pages/Profile";
import Admin from "./pages/Admin";
import NoMatch from "./pages/404";

import "react-toastify/dist/ReactToastify.css";

import "./App.scss"

import { pdfjs } from "react-pdf";
import logout from "./pages/Logout/index.jsx";
import Logout from "./pages/Logout/index.jsx";
pdfjs.GlobalWorkerOptions.workerSrc = new URL(
  'pdfjs-dist/build/pdf.worker.min.mjs',
  import.meta.url,
).toString(); 

export default function App() {

  useEffect(() => {
    apiService.checkLogin()
  }, [])

  return (
    <>
      <AuthProvider>
        <Router>
          <Navigationbar />
            <Switch>
              <PrivateRoute exact path="/" component={Home} />
              <PrivateRoute path="/documents" component={Documents} />
              <PrivateRoute path="/connect" component={Connect} />
              <PrivateRoute path="/integrations" component={Integrations} />
              <PrivateRoute path="/profile" component={Profile} />
              <PrivateRoute path="/admin" roles={[Role.Admin]} component={Admin} />

              <Route path="/login" component={Login} />
              <Route path="/logout" component={Logout} />
              <Route component={NoMatch} />
            </Switch>
        </Router>
      </AuthProvider>
      <ToastContainer autoClose={2000} />
    </>
  );
}
