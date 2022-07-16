import React from "react";
import { Nav, Navbar, Button, NavDropdown, Container } from "react-bootstrap";
import { logout } from "../common/actions";
import { useAuthState } from "../common/useAuthContext";
import { NavLink } from "react-router-dom";

const NavigationBar = () => {
  const { state:{user}, dispatch } = useAuthState();

  function isAdmin(user) {
    return user && user.Roles && user.Roles[0] === "Admin";
  }

  function handleLogout(e) {
    logout(dispatch);
  }

  return (
    <Navbar bg="dark" variant="dark">
      <Container fluid={true}>
      <Navbar.Brand href="/">rmfakecloud</Navbar.Brand>
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
                  <Nav.Link as={NavLink} to="/users">Users</Nav.Link>
                </Nav.Item>
              )}
              <Nav.Item>
                <Nav.Link as={NavLink} to="/generatecode">Code</Nav.Link>
              </Nav.Item>
              </Nav>
              <Nav className="ms-auto">
              <NavDropdown id="userMenu" title={user.UserID} alignRight align="end" style={{'marginRight':50}}>
                {user.scopes === "sync15" && (<NavDropdown.Header>Using sync 15</NavDropdown.Header>)}
                <NavDropdown.Item as={NavLink} to="/resetPassword">Reset Password</NavDropdown.Item>
                <NavDropdown.Divider />
                <NavDropdown.Item as={Button} onClick={handleLogout}>Log out </NavDropdown.Item>
              </NavDropdown>
             </Nav>
          </Navbar.Collapse>
        </>
      )}
    </Container>
    </Navbar>
  );
};

export default NavigationBar;
