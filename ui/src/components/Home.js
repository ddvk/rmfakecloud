import React from "react";
import Container from "react-bootstrap/Container";

const Home = () => {
  return (
    <>
    <Container>
    <div>
        <h3>Welcome to your own rm cloud</h3>
        <h3>About</h3>
        <p>This is still a work in progress</p>
        <h3>TODO</h3>
        <p>
          <ul>
          <li>token is expiration is not checked, so you'll be logged out when that happens</li>
          <li>you can you rmapi (https://github.com/juruen/rmapi) for managing files</li>
          </ul>
          </p>
    </div>
  </Container>
    </>
  );
};

export default Home;
