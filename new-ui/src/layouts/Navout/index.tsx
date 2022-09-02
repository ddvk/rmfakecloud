import { Component } from 'react'
import { Outlet } from 'react-router-dom'

import Navbar from './navbar'

export default class Navout extends Component {
  render() {
    return (
      <>
        <Navbar />
        <Outlet />
      </>
    )
  }
}
