import React from "react";
import Container from "react-bootstrap/Container";
import Stack from "react-bootstrap/Stack";
import { useAuthState } from "../../common/useAuthContext";

import UserList from "./UserList";
import ResetPassword from "./ResetPassword";

const Home = () => {
  const { state: { user } } = useAuthState();

  function isAdmin(user) {
    return user && user.Roles && user.Roles[0] === "Admin";
  }

  return (
    <Container fluid>
      <Stack>
        <div>
          {user.scopes === "sync15" && (<span>Using sync 15</span>)}
        </div>
        <div>
          <ResetPassword />
        </div>
        { isAdmin(user) && <div>
          <UserList />
        </div>}
      </Stack>
    </Container>
  );
};

export default Home;
