import Container from "react-bootstrap/Container";
import Stack from "react-bootstrap/Stack";
import UserList from "./UserList";
import ServerSettings from "./ServerSettings";

const Home = () => {
  return (
    <Container fluid>
      <Stack gap={4}>
        <ServerSettings />
        <UserList />
      </Stack>
    </Container>
  );
};

export default Home;
