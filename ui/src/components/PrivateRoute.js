import React from "react";
import { Route, Redirect } from "react-router-dom";

import { useAuthState } from "../hooks/useAuthContext";

export const PrivateRoute = ({ component: Component, ...rest }) => {
  const userDetails = useAuthState(); //read the values of loading and errorMessage from context
  return (
    <Route
      {...rest}
      render={(props) => {
        if (!userDetails.user) {
          // not logged in so redirect to login page with the return url
          return (
            <Redirect
              to={{ pathname: "/login", state: { from: props.location } }}
            />
          );
        }

        // // check if route is restricted by role
        // if (roles && roles.indexOf(currentUser.role) === -1) {
        //   // role not authorised so redirect to home page
        //   return <Redirect to={{ pathname: "/" }} />;
        // }

        // authorised so return component
        return <Component {...props} />;
      }}
    />
  );
};
