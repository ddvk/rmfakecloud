import "bootstrap/dist/css/bootstrap.min.css";
import React from "react";
import Layout from "./components/Layout";
import Navigationbar from "./components/NavigationBar";
import FileList from "./components/FileList";
import FileListFunctional from "./components/FileListFunction";
import NoMatch from "./components/NoMatch";
import Home from "./components/Home";
import { BrowserRouter as Router, Route, Switch } from "react-router-dom";
import { AuthProvider } from "./components/Login/useAuthContext";
import Login from "./components/Login/Login";
import { PrivateRoute } from "./components/PrivateRoute";

export default function App() {
  return (
    <AuthProvider>
      <Router>
        <Navigationbar />
        <Layout>
          <Switch>
            <PrivateRoute exact path="/" component={Home} />
            <PrivateRoute path="/fileList" component={FileList} />
            <PrivateRoute
              path="/fileListFunctional"
              component={FileListFunctional}
            />
            <Route path="/login" component={Login} />
            <Route component={NoMatch} />
          </Switch>
        </Layout>
      </Router>
    </AuthProvider>
  );
}
