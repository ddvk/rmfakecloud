import React, { useEffect } from "react";
import { useHistory } from "react-router-dom";

import apiService from "../../services/api.service";
import { useAuthState } from "../../common/useAuthContext";

const TIMEOUT_MS = 10000;

const OidcCallback = () => {
  const { dispatch } = useAuthState();
  const history = useHistory();

  useEffect(() => {
    dispatch({ type: "REQUEST_LOGIN" });

    const timeout = setTimeout(() => {
      dispatch({ type: "LOGIN_ERROR", error: "OIDC login timed out" });
      history.replace("/login");
    }, TIMEOUT_MS);

    apiService
      .me()
      .then((user) => {
        clearTimeout(timeout);
        dispatch({ type: "LOGIN_SUCCESS", payload: { user } });
        history.replace("/documents");
      })
      .catch(() => {
        clearTimeout(timeout);
        dispatch({ type: "LOGIN_ERROR", error: "OIDC login failed" });
        history.replace("/login");
      });

    return () => clearTimeout(timeout);
  }, [dispatch, history]);

  return (
    <div style={{ textAlign: "center", marginTop: "2rem" }}>
      <span>Completing login…</span>
    </div>
  );
};

export default OidcCallback;
