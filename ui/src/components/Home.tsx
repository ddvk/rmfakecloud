import React from "react";
import Container from "react-bootstrap/Container";

const Home = () => {
  return (
    <Container>
      <div>
        <h3>Welcome to your own rm cloud</h3>
        <h3>About</h3>
        <p>This is still a work in progress</p>
        <h3>TIPS</h3>
        <ul>
          <li>
            you can use rmapi (https://github.com/juruen/rmapi) for managing
            files, just specify the URL of your instance with the RMAPI_HOST
            variable (RMAPI_HOST=https//rmfakeclud.example.com rmapi)
          </li>
        </ul>
      </div>
    </Container>
  );
};

export default Home;
