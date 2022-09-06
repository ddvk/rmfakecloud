import { Component } from 'react'
import { Outlet } from 'react-router-dom'
import { ToastContainer } from 'react-toastify'

import Navbar from './navbar'

export default class Navout extends Component {
  render() {
    return (
      <>
        <Navbar />
        <Outlet />
        <ToastContainer />
      </>
    )
  }
}
