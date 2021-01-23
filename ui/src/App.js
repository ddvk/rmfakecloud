import "bootstrap/dist/css/bootstrap.min.css";
import React from "react";
import Row from "react-bootstrap/Row";
import Layout from "./components/Layout";
import Navigationbar from "./components/NavigationBar";
import FileList from "./components/FileList";
import FileListFunctional from "./components/FileListFunction";
import NoMatch from "./components/NoMatch";
import Home from "./components/Home";
import { BrowserRouter as Router, Route, Switch } from "react-router-dom";

export default function App() {
  return (
    <>
      <Navigationbar />
      <Layout>
        <Router>
          <Switch>
            <Route exact path="/" component={Home} />
            <Route path="/fileList" component={FileList} />
            <Route path="/fileListFunctional" component={FileListFunctional} />
            <Route component={NoMatch} />
            <Row>
              <div className="flex-column">
                <FileListFunctional />
                <FileList />
              </div>
            </Row>
          </Switch>
        </Router>
      </Layout>
    </>
  );
}
