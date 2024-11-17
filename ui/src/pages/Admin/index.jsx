import Container from "react-bootstrap/Container";
import Stack from "react-bootstrap/Stack";
import UserList from "./UserList";

const Home = () => {
  return (
    <Container fluid>
      <Stack>
          <UserList />
      </Stack>
    </Container>
  );
};

export default Home;
