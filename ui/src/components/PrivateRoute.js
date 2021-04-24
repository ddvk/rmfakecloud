import React from "react";
import { Route, Redirect } from "react-router-dom";
import { logout } from "./Login/actions";
import { useAuthState, useAuthDispatch } from "../hooks/useAuthContext";

export const PrivateRoute = ({ component: Component, roles, ...rest }) => {
  const { user } = useAuthState(); //read the values of loading and errorMessage from context
  const authDispatch = useAuthDispatch();
  return (
    <Route
      {...rest}
      render={(props) => {
        if (!user) {
          // not logged in so redirect to login page with the return url
          return (
            <Redirect
              to={{ pathname: "/login", state: { from: props.location } }}
            />
          );
        }

        debugger;

        // check if route is restricted by role
        if (roles && user.Roles && roles.indexOf(user.Roles[0]) === -1) {
          // role not authorised ==> logout
          logout(authDispatch);
          return <Redirect to={{ pathname: "/login" }} />;
        }

        // authorised so return component
        return <Component {...props} />;
      }}
    />
  );
};
