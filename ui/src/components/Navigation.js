import React from "react";
import { Nav, Navbar, Button, NavDropdown, Container } from "react-bootstrap";
import { logout } from "../common/actions";
import { useAuthState } from "../common/useAuthContext";
import { NavLink } from "react-router-dom";

const NavigationBar = () => {
  const { state:{user}, dispatch } = useAuthState();

  function handleLogout(e) {
    logout(dispatch);
  }

  return (
    <Navbar variant="dark" className="sticky-top" style={{ backdropFilter: 'blur(12px)', backgroundColor: 'rgba(255, 255, 255, .1)' }}>
      <Container>
        <Navbar.Brand>
          <Nav.Link as={NavLink} to="/">
            rmfakecloud
          </Nav.Link>
        </Navbar.Brand>
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
              </Nav>
              <Nav className="ms-auto">
                <NavDropdown id="userMenu" title={user.UserID} align="end" menuVariant="dark">
                  <NavDropdown.Item as={NavLink} to="/connect">Connect Device</NavDropdown.Item>
                  <NavDropdown.Item as={NavLink} to="/admin">Admin</NavDropdown.Item>
                  <NavDropdown.Item as={NavLink} to="/about">About</NavDropdown.Item>
                  <NavDropdown.Divider />
                  <NavDropdown.Item as={Button} onClick={handleLogout}>Log out</NavDropdown.Item>
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
