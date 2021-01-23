import React from "react";
import { Nav, Navbar } from "react-bootstrap";

const NavigationBar = () => {
  return (
    <Navbar bg="dark" variant="dark">
      <Navbar.Brand href="/">freeMarkable</Navbar.Brand>
      <Navbar.Toggle />
      <Navbar.Collapse>
        <Nav>
          {/* // activeKey="/home"
          // onSelect={(k) => console.log(k)}
          // className="mr-auto" */}
          <Nav.Item>
            <Nav.Link href="/">Home</Nav.Link>
          </Nav.Item>
          <Nav.Item>
            <Nav.Link href="/filelist">FileList</Nav.Link>
          </Nav.Item>
          <Nav.Item>
            <Nav.Link href="/filelistFunctional">FileList Functional</Nav.Link>
          </Nav.Item>
        </Nav>
      </Navbar.Collapse>
    </Navbar>
  );
};

export default NavigationBar;
