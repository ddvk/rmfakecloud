import {useEffect} from "react";
import { BrowserRouter as Router, Route, Switch } from "react-router-dom";
import { ToastContainer } from 'react-toastify';

import apiService from "./services/api.service";
import { AuthProvider } from "./common/useAuthContext";
import Role from "./common/Role";
import { PrivateRoute } from "./components/PrivateRoute";
import Navigationbar from "./components/Navigation";
import SuBanner from "./components/SuBanner";

import Login from "./pages/Login";
import Home from "./pages/Home";
import Connect from "./pages/Connect";
import Documents from "./pages/Documents";
import ViewPdf from "./pages/ViewPdf";
import Integrations from "./pages/Integrations";
import Profile from "./pages/Profile";
import Admin from "./pages/Admin";
import Logout from "./pages/Logout";
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
          <div style={{display: "flex", flexDirection: "column", height: "100%"}}>
            <Navigationbar />
            <SuBanner />
            <div style={{flex: "1 1 auto", minHeight: 0, overflow: "hidden"}}>
              <Switch>
                <PrivateRoute exact path="/" component={Home} />
                <PrivateRoute path="/view-pdf/:docId" component={ViewPdf} />
                <PrivateRoute path="/documents/:itemId?" component={Documents} />
                <PrivateRoute path="/connect" component={Connect} />
                <PrivateRoute path="/pair/app" component={Connect} />
                <PrivateRoute path="/pair" component={Connect} />
                <PrivateRoute path="/integrations" component={Integrations} />
                <PrivateRoute path="/profile" component={Profile} />
                <PrivateRoute path="/admin" roles={[Role.Admin]} component={Admin} />
                <Route path="/logout" component={Logout} />

                <Route path="/login" component={Login} />
                <Route component={NoMatch} />
              </Switch>
            </div>
          </div>
        </Router>
      </AuthProvider>
      <ToastContainer autoClose={2000} />
    </>
  );
}
