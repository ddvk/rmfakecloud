import {Component} from "react";
import {Outlet} from "react-router-dom";

export default class Navout extends Component {
  render() {
    return (
      <>
        <Outlet />
      </>
    )
  }
}
