import React from "react";
import Container from "react-bootstrap/Container";
import Stack from "react-bootstrap/Stack";
import { useAuthState } from "../../common/useAuthContext";

import UserList from "./UserList";
import ResetPassword from "./ResetPassword";

const Home = () => {
  const { state: { user } } = useAuthState();

  const style = { margin: '2em 0' }

  function isAdmin(user: any) {
    return user && user.Roles && user.Roles[0] === "Admin";
  }

  return (
    <Container>
      <Stack>
        <div style={style}>
          {user.scopes === "sync15" && (<span>Using sync 15</span>)}
        </div>
        <div style={style}>
          <ResetPassword />
        </div>
        { isAdmin(user) && <div style={style}>
          <UserList />
        </div>}
      </Stack>
    </Container>
  );
};

export default Home;
