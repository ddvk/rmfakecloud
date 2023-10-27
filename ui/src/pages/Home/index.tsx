import React from "react";
import Upload from "./Upload";
import Container from "react-bootstrap/Container";

const fileUploaded = () => {
  //callback
};

const Home = () => {
  return (
    <Container>
      <Upload className="m-5" filesUploaded={fileUploaded} uploadFolder="" />

      <h3>Recent files</h3>
      <p>TODO: show recent files in a grid with preview images</p>
    </Container>
  );
};

export default Home;
