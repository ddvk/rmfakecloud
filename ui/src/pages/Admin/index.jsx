import Container from "react-bootstrap/Container";
import Stack from "react-bootstrap/Stack";
import { useAuthState } from "../../common/useAuthContext";

import UserList from "./UserList";

const Home = () => {
  const { state: { user } } = useAuthState();

  function isAdmin(user) {
    return user && user.Roles && user.Roles[0] === "Admin";
  }

  return (
    <Container fluid>
      <Stack>
          <UserList />
      </Stack>
    </Container>
  );
};

export default Home;
