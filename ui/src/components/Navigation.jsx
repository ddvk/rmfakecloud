import React from "react";
import { Nav, Navbar, NavDropdown, Container } from "react-bootstrap";
import { useAuthState } from "../common/useAuthContext";
import { NavLink } from "react-router-dom";

const NavigationBar = () => {
  const { state:{user} } = useAuthState();

  function isAdmin() {
    return user && user.Roles && user.Roles[0] === "Admin";
  }
  return (
    <Navbar className="sticky-top">
      <Container fluid>
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
                <Nav.Item>
                  <Nav.Link as={NavLink} to="/integrations">
                    Integrations
                  </Nav.Link>
                </Nav.Item>
                <Nav.Item>
                  <Nav.Link as={NavLink} to="/connect">
                    Connect
                  </Nav.Link>
                </Nav.Item>
				{ isAdmin() &&

					<Nav.Item>
					  <Nav.Link as={NavLink} to="/admin">
						Admin	
					  </Nav.Link>
					</Nav.Item>
				}
              </Nav>
              <Nav className="ms-auto">
                <NavDropdown id="userMenu" title={user.UserID} align="end">
                  <NavDropdown.Item as={NavLink} to="/profile">Profile</NavDropdown.Item>
                  <NavDropdown.Divider />
                  <NavDropdown.Item as={NavLink} to="/logout">Log out</NavDropdown.Item>
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
