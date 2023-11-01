import React from "react";
import Upload from "./Upload";
import Container from "react-bootstrap/Container";
import { Nav } from "react-bootstrap";
import { NavLink } from "react-router-dom";

const fileUploaded = () => {
  //callback
};

const Home = () => {
  return (
    <Container>
      <Upload className="m-5" filesUploaded={fileUploaded} uploadFolder="" />

      <h3>My Files</h3>
      <Nav.Link as={NavLink} to="/documents">
        Go to Documents
      </Nav.Link>
      {/*
      <h3>Recent files</h3>
      <p>TODO: show recent files in a grid with preview images</p>
        */}
    </Container>
  );
};

export default Home;
