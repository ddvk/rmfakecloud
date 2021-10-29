import React, { useState } from "react";
import { useAuthState } from "../../common/useAuthContext";
import { loginUser } from "../../common/actions";
import styles from "./Login.module.css";
import { useHistory } from "react-router";

const Login = () => {
  let history = useHistory();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const { state, dispatch } = useAuthState(); //read the values of loading and errorMessage from context
  const { errorMessage, loading } = state;

  const handleLogin = async (e) => {
    e.preventDefault();

    let payload = { email, password };
    try {
      await loginUser(dispatch, payload);
      history.push("/"); //TODO: usenavigate or return redirect
    } catch (error) {
      console.log(error);
    }
  };

  return (
    <div className={styles.container}>
      <div style={{ width: 200 }}>
        {errorMessage ? <p className={styles.error}>{errorMessage}</p> : null}
        <form>
          <div className={styles.loginForm}>
            <div className={styles.loginFormItem}>
              <label htmlFor="email">Username</label>
              <input
                type="text"
                id="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                disabled={loading}
                autoFocus
              />
            </div>
            <div className={styles.loginFormItem}>
              <label htmlFor="password">Password</label>
              <input
                type="password"
                id="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={loading}
              />
            </div>
          </div>
          <button onClick={handleLogin} disabled={loading}>
            login
          </button>
        </form>
      </div>
    </div>
  );
};

export default Login;
