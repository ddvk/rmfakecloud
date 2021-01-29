import React, { useState } from "react";
import { useAuthDispatch, useAuthState } from "./useAuthContext";
import { loginUser } from "./actions";
import styles from "./Login.module.css";

const Login = ({ history }) => {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const dispatch = useAuthDispatch(); //get the dispatch method from the useDispatch custom hook
  const { loading, errorMessage } = useAuthState(); //read the values of loading and errorMessage from context

  const handleLogin = async (e) => {
    e.preventDefault();

    let payload = { email, password };
    try {
      let response = await loginUser(dispatch, payload);
      if (!response.user) return;

      history.push("/"); //TODO: usenavigate or return redirect
    } catch (error) {
      console.log(error);
    }
  };

  return (
    <div className={styles.container}>
      <div className={{ width: 200 }}>
        <h1>Login Page</h1>
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
