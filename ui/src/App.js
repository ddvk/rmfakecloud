import React from "react";

import Layout from "./components/Layout/Layout";
import Navigationbar from "./components/Navigation/NavigationBar";
import Login from "./components/Login/Login";
import UserList from "./components/User/UserList";
import UserProfile from "./components/User/UserProfile";
import Home from "./components/Home";
import FileList from "./components/File/FileList";
import FileListFunctional from "./components/File/FileListFunction";
import NoMatch from "./components/NoMatch";

import { BrowserRouter as Router, Route, Switch } from "react-router-dom";
import { AuthProvider } from "./hooks/useAuthContext";
import { PrivateRoute } from "./components/PrivateRoute";
import CodeGenerator from "./components/User/CodeGenerator";

import "bootstrap/dist/css/bootstrap.min.css";

export default function App() {
  return (
    <AuthProvider>
      <Router>
        <Navigationbar />
        <Layout>
          <Switch>
            <PrivateRoute exact path="/" component={Home} />
            <PrivateRoute path="/documents" component={FileListFunctional} />
            <PrivateRoute path="/userList/:userid" component={UserProfile} />
            <PrivateRoute path="/userList" component={UserList} />
            <PrivateRoute path="/generatecode" component={CodeGenerator} />

            <Route path="/login" component={Login} />
            <Route component={NoMatch} />
          </Switch>
        </Layout>
      </Router>
    </AuthProvider>
  );
}
