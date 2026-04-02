import { useState } from "react";
import { Alert, Button, Table } from "react-bootstrap";

import useFetch from "../../hooks/useFetch";
import Spinner from "../../components/Spinner";
import { useAuthState } from "../../common/useAuthContext";
import apiService from "../../services/api.service";

export default function SudoPage() {
  const { state: { user }, dispatch } = useAuthState();
  const { data: userList, error, loading } = useFetch("users");
  const [busy, setBusy] = useState("");
  const [msg, setMsg] = useState("");

  if (loading) return <Spinner />;
  if (error) {
    return (
      <Alert variant="danger">
        <Alert.Heading>Unable to load users</Alert.Heading>
        {`Error ${error.status || ""} ${error.statusText || String(error)}`}
      </Alert>
    );
  }

  async function become(userid) {
    setBusy(userid);
    setMsg("");
    try {
      const sudoUser = await apiService.sudoAs(userid);
      dispatch({ type: "LOGIN_SUCCESS", payload: { user: sudoUser } });
      setMsg(`Now acting as ${sudoUser.UserID}`);
    } catch (e) {
      setMsg(e.message || String(e));
    } finally {
      setBusy("");
    }
  }

  async function returnSudo() {
    setBusy("__return__");
    setMsg("");
    try {
      const sudoUser = await apiService.returnFromSudo();
      dispatch({ type: "LOGIN_SUCCESS", payload: { user: sudoUser } });
      setMsg(`Returned to ${sudoUser.UserID}`);
    } catch (e) {
      setMsg(e.message || String(e));
    } finally {
      setBusy("");
    }
  }

  return (
    <>
      <h3>Sudo</h3>
      <p className="text-muted">
        Admin-only session switching. Choose a user to act as them.
      </p>
      {user?.SudoBy && (
        <Alert variant="warning">
          Acting as <strong>{user.UserID}</strong> (sudo by {user.SudoBy})
          <div className="mt-2">
            <Button size="sm" variant="outline-dark" disabled={busy !== ""} onClick={returnSudo}>
              {busy === "__return__" ? "Returning..." : "Return to original admin"}
            </Button>
          </div>
        </Alert>
      )}
      {msg && <Alert variant="info">{msg}</Alert>}
      <Table striped bordered hover>
        <thead>
          <tr>
            <th>User</th>
            <th>Email</th>
            <th>Role</th>
            <th />
          </tr>
        </thead>
        <tbody>
          {(userList || []).map((u) => (
            <tr key={u.userid}>
              <td>{u.userid}</td>
              <td>{u.email}</td>
              <td>{u.isAdmin ? "admin" : "user"}</td>
              <td>
                <Button size="sm" disabled={busy !== ""} onClick={() => become(u.userid)}>
                  {busy === u.userid ? "Switching..." : "Sudo as"}
                </Button>
              </td>
            </tr>
          ))}
        </tbody>
      </Table>
    </>
  );
}

