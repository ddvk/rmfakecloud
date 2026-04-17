import { useEffect, useState } from "react";
import Button from "react-bootstrap/Button";
import Container from "react-bootstrap/Container";
import ListGroup from "react-bootstrap/ListGroup";

import apiservice from "../services/api.service";
import { useAuthState } from "../common/useAuthContext";

export default function PasscodeResets() {
  const { state: { user } } = useAuthState();
  const [resets, setResets] = useState([]);
  const [error, setError] = useState("");

  const load = () => {
    apiservice
      .listPasscodeResets()
      .then((list) => setResets(list || []))
      .catch((e) => setError(e.toString()));
  };

  useEffect(() => {
    if (!user) {
      setResets([]);
      return;
    }
    load();
    const id = setInterval(load, 5000);
    return () => clearInterval(id);
  }, [user]);

  const approve = (uuid) => {
    apiservice
      .approvePasscodeReset(uuid)
      .then(load)
      .catch((e) => setError(e.toString()));
  };

  const dismiss = (uuid) => {
    apiservice
      .dismissPasscodeReset(uuid)
      .then(load)
      .catch((e) => setError(e.toString()));
  };

  if (!user) {
    return null;
  }
  if (error) {
    return (
      <Container fluid className="pt-3">
        <div className="alert alert-danger mb-0">{error}</div>
      </Container>
    );
  }
  if (resets.length === 0) {
    return null;
  }

  return (
    <Container fluid className="pt-3">
      <ListGroup className="mb-0">
        {resets.map((r) => (
          <ListGroup.Item
            key={r.RequestID}
            className="d-flex justify-content-between align-items-center"
          >
            <div>
              <div><strong>Passcode reset requested</strong> — {r.DeviceName || "device"}</div>
              <small className="text-muted">
                {r.DeviceID} &middot; {new Date(r.Created).toLocaleString()}
              </small>
            </div>
            <div className="d-flex gap-2">
              <Button variant="outline-secondary" onClick={() => dismiss(r.RequestID)}>
                Dismiss
              </Button>
              <Button variant="primary" onClick={() => approve(r.RequestID)}>
                Approve
              </Button>
            </div>
          </ListGroup.Item>
        ))}
      </ListGroup>
    </Container>
  );
}
