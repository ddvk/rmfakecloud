import Container from "react-bootstrap/Container";
import Stack from "react-bootstrap/Stack";
import { useAuthState } from "../../common/useAuthContext";

import ResetPassword from "./ResetPassword";

const Home = () => {
  const { state: { user } } = useAuthState();
  return (
    <Container fluid>
      <Stack>
        <div>
          {user.scopes === "sync15" && (<span>Using sync 15</span>)}
        </div>
        <div>
          <ResetPassword />
        </div>
      </Stack>
    </Container>
  );
};

export default Home;
