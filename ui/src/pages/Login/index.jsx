import React, {useEffect, useState} from "react";
import {useHistory} from "react-router-dom";
import {Button, Form} from "react-bootstrap";
import Spinner from "../../components/Spinner";

import {useAuthState} from "../../common/useAuthContext";
import {loginUser, oidcAuth, oidcCallback} from "../../common/actions";

import styles from "./Login.module.scss";
import apiService from "../../services/api.service.js";

const Login = () => {
  let history = useHistory();

  const [oidcInfo, setOidcInfo] = useState(null);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  const { state, dispatch } = useAuthState(); //read the values of loading and errorMessage from context
  const { errorMessage, loading: waiting } = state;

  useEffect(() => {
    const init = async () => {
      try {
        const searchParams = new URLSearchParams(window.location.search)

        if (searchParams.has("code")) {
          let payload = {code: searchParams.get("code")};
          try {
            await oidcCallback(dispatch, payload)
            history.push("/documents"); //TODO: usenavigate or return redirect
            return
          } catch (error) {
            console.log(error);
          }
        }

        const json = await apiService.oidcInfo()

        if (json != null && json.only) {
          await handleOidc(null)
          return
        }

        setOidcInfo(json);
      } catch (error) {
        dispatch({type: "LOGIN_ERROR", error: "Can't login: " + error.message});
      }
    }

   init()
  }, []);

  const handleLogin = async (e) => {
    e.preventDefault();

    let payload = { email: username, password };
    try {
      await loginUser(dispatch, payload);
      history.push("/documents"); //TODO: usenavigate or return redirect
    } catch (error) {
      console.log(error);
    }
  };
  
  const handleOidc = async (e) => {
    if (e != null) e.preventDefault()
    
    try {
      await oidcAuth(dispatch).then((url) =>  {
        if (url !== undefined) window.location.href = url
      })
    } catch (error) {
      console.log(error);
    }
  }

  return (
    <div className={styles.container}>
      <div className={styles.formContainer}>
        {errorMessage ? <p className={styles.error}>{errorMessage}</p> : null}
        {oidcInfo == null ? <Spinner/> : <>
          <Form>
            <Form.Group className="mb-3">
              <Form.Label htmlFor="username">Username</Form.Label>
              <Form.Control
                id="username"
                value={username}
                autoFocus
                onChange={(e) => setUsername(e.target.value)}
                disabled={waiting}
                placeholder="Username"
                autoComplete="username"
              />
            </Form.Group>

            <Form.Group className="mb-3">
              <Form.Label htmlFor="password">Password</Form.Label>
              <Form.Control
                type="password"
                id="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={waiting}
                placeholder="Password"
                autoComplete="current-password"
              />
            </Form.Group>

            <Button type="submit" onClick={handleLogin} disabled={waiting}>
              Login
            </Button>
          </Form>
        </>}
        {oidcInfo != null && oidcInfo.enabled ? <>
          <hr/>
          <Button onClick={handleOidc} disabled={waiting}>
            {oidcInfo.label}
          </Button>
        </> : null}
      </div>
    </div>
  );
};

export default Login;
