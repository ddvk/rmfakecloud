import React, { useEffect, useState } from "react";
import { useHistory } from "react-router-dom";
import { Button, Form } from "react-bootstrap";
import { startAuthentication, browserSupportsWebAuthn } from "@simplewebauthn/browser";

import { useAuthState } from "../../common/useAuthContext";
import { loginUser } from "../../common/actions";
import apiService from "../../services/api.service";

import styles from "./Login.module.scss";

const Login = () => {
  let history = useHistory();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [passkeyEnabled, setPasskeyEnabled] = useState(false);
  const [passkeyBusy, setPasskeyBusy] = useState(false);

  const { state, dispatch } = useAuthState();
  const { errorMessage, loading } = state;

  useEffect(() => {
    let cancelled = false;
    (async () => {
      if (!browserSupportsWebAuthn()) return;
      try {
        const st = await apiService.webAuthnStatus();
        if (!cancelled) setPasskeyEnabled(Boolean(st && st.enabled));
      } catch (_) {
        if (!cancelled) setPasskeyEnabled(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const handleLogin = async (e) => {
    e.preventDefault();

    let payload = { email: username, password };
    try {
      await loginUser(dispatch, payload);
      history.push("/documents");
    } catch (error) {
      console.log(error);
    }
  };

  const handlePasskeyLogin = async (e) => {
    e.preventDefault();
    setPasskeyBusy(true);
    dispatch({ type: "REQUEST_LOGIN" });
    try {
      const begin = await apiService.webAuthnLoginBegin();
      const credential = await startAuthentication({ optionsJSON: begin.publicKey });
      const user = await apiService.webAuthnLoginFinish(begin.sessionId, credential);
      dispatch({
        type: "LOGIN_SUCCESS",
        payload: { user },
      });
      history.push("/documents");
    } catch (error) {
      dispatch({
        type: "LOGIN_ERROR",
        error: "Passkey login failed: " + (error.message || String(error)),
      });
    } finally {
      setPasskeyBusy(false);
    }
  };

  return (
    <div className={styles.container}>
      <div className={styles.formContainer}>
        {errorMessage ? <p className={styles.error}>{errorMessage}</p> : null}

        <Form>
          <Form.Group className="mb-3">
            <Form.Label htmlFor="username">Username</Form.Label>
            <Form.Control
              id="username"
              value={username}
              autoFocus
              onChange={(e) => setUsername(e.target.value)}
              disabled={loading || passkeyBusy}
              placeholder="Username"
              autoComplete="username webauthn"
            />
          </Form.Group>

          <Form.Group className="mb-3">
            <Form.Label htmlFor="password">Password</Form.Label>
            <Form.Control
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={loading || passkeyBusy}
              placeholder="Password"
              autoComplete="current-password"
            />
          </Form.Group>

          <Button type="submit" onClick={handleLogin} disabled={loading || passkeyBusy}>
            Login
          </Button>
          {passkeyEnabled ? (
            <Button
              type="button"
              variant="outline-secondary"
              className="ms-2"
              onClick={handlePasskeyLogin}
              disabled={loading || passkeyBusy}
            >
              Sign in with passkey
            </Button>
          ) : null}
        </Form>
      </div>
    </div>
  );
};

export default Login;
