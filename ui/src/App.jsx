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
import Admin from "./pages/Admin";
import NoMatch from "./pages/404";

import "react-toastify/dist/ReactToastify.css";

import "./App.scss"

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
              <PrivateRoute path="/admin" roles={[Role.Admin]} component={Admin} />

              <Route path="/login" component={Login} />
              <Route component={NoMatch} />
            </Switch>
        </Router>
      </AuthProvider>
      <ToastContainer autoClose={2000} />
    </>
  );
}
