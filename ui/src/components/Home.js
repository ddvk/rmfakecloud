import React from "react";
import Jumbo from "react-bootstrap/Jumbotron";
import Container from "react-bootstrap/Container";
import booksImage from "../assets/books.jpeg";

const Home = () => {
  return (
    <>
      <Jumbo
        fluid
        style={{
          //background: `url(${booksImage})`,
          backgroundSize: "cover",
          color: "white",
        }}
      >
        <Container>
          <h1>Welcome to freeMarkable Cloud.</h1>
        </Container>
      </Jumbo>
    </>
  );
};

export default Home;
