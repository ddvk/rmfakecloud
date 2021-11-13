import React, {useEffect} from "react";

import Navigationbar from "./components/Navigation";
import Login from "./components/Login";
import UserList from "./components/UserList";
import UserProfile from "./components/UserProfile";
import Home from "./components/Home";
import Documents from "./components/Documents";
import NoMatch from "./components/NoMatch";

import { BrowserRouter as Router, Route, Switch } from "react-router-dom";
import { AuthProvider } from "./common/useAuthContext";
import { PrivateRoute } from "./components/PrivateRoute";
import CodeGenerator from "./components/CodeGenerator";
import ResetPassword from "./components/ResetPassword";
import Role from "./common/Role";
import apiService from "./services/api.service";

import "bootstrap/dist/css/bootstrap.min.css";

export default function App() {
  useEffect(() => {
    apiService.checkLogin()
  }, [])
  return (
    <AuthProvider>
      <Router>
        <Navigationbar />
          <Switch>
            <PrivateRoute exact path="/" component={Home} />
            <PrivateRoute path="/documents" component={Documents} />
            <PrivateRoute path="/generatecode" component={CodeGenerator} />
            <PrivateRoute path="/resetPassword" component={ResetPassword} />
            <PrivateRoute path="/userList/:userid" component={UserProfile} /> 
            <PrivateRoute path="/userList" roles={[Role.Admin]} component={UserList} />
            <Route path="/login" component={Login} />
            <Route component={NoMatch} />
          </Switch>
      </Router>
    </AuthProvider>
  );
}
