import Container from "react-bootstrap/Container";
import Stack from "react-bootstrap/Stack";
import { useAuthState } from "../../common/useAuthContext";

import ResetPassword from "./ResetPassword";
import RegisteredDevices from "./RegisteredDevices";

const Home = () => {
  const { state: { user } } = useAuthState();
  return (
    <Container fluid>
      <Stack>
        <div>
          {user.scopes === "sync15" && (<span>Using sync 15</span>)}
        </div>
        <div>
          <RegisteredDevices />
        </div>
        <div>
          <ResetPassword />
        </div>
      </Stack>
    </Container>
  );
};

export default Home;
