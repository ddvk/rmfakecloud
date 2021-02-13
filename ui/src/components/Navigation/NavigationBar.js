import React from "react";
import { Nav, Navbar, Button, NavDropdown } from "react-bootstrap";
import { logout } from "../Login/actions";
import { useAuthDispatch, useAuthState } from "../../hooks/useAuthContext";
import { NavLink } from "react-router-dom";
import styles from "./NavigationBar.module.css";
import { FaUser } from "react-icons/fa";

const NavigationBar = () => {
  const authDispatch = useAuthDispatch();
  const { user } = useAuthState();

  function isAdmin(user) {
    return user && user.Roles && user.Roles[0] === "Admin";
  }

  function handleLogout(e) {
    logout(authDispatch);
  }

  return (
    <Navbar bg="dark" variant="dark">
      <Navbar.Brand href="/">freeMarkable</Navbar.Brand>
      <Navbar.Toggle />
      {user && (
        <>
          <Navbar.Collapse>
            <Nav>
              {" "}
              <Nav.Item>
                <Nav.Link as={NavLink} to="/documents">
                  Documents
                </Nav.Link>
              </Nav.Item>
              {isAdmin(user) && (
                <Nav.Item>
                  <Nav.Link as={NavLink} to="/userList">
                    Users
                  </Nav.Link>
                </Nav.Item>
              )}
              <Nav.Item>
                <Nav.Link as={NavLink} to="/generatecode">
                  Code
                </Nav.Link>
              </Nav.Item>
            </Nav>
          </Navbar.Collapse>
          <Navbar.Collapse>
            <Nav className="ml-auto">
              <NavDropdown
                alignRight
                title={
                  <span>
                    <FaUser size={30} />
                  </span>
                }
              >
                <NavDropdown.Item as={NavLink} to="/profile">
                  Profile
                </NavDropdown.Item>
                <NavDropdown.Item as={NavLink} to="/resetPassword">
                  Reset Password
                </NavDropdown.Item>
                <NavDropdown.Divider />
                <NavDropdown.Item as={Button} onClick={handleLogout}>
                  Log out
                </NavDropdown.Item>
              </NavDropdown>
            </Nav>
          </Navbar.Collapse>
        </>
      )}
    </Navbar>
  );
};

export default NavigationBar;
