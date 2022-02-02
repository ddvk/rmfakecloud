import React from "react";
import { Route, Redirect } from "react-router-dom";
import { logout } from "../common/actions";
import { useAuthState } from "../common/useAuthContext";

type RouteProp = {
  component: React.ComponentType<any>;
  roles?: string[];
  exact?: boolean;
  path: string;
};

export const PrivateRoute = ({
  component: Component,
  roles,
  ...rest
}: RouteProp) => {
  const { state:{user}, dispatch } = useAuthState(); //read the values of loading and errorMessage from context
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

        // check if route is restricted by role
        if (roles && user.Roles && roles.indexOf(user.Roles[0]) === -1) {
          // role not authorised ==> logout
          logout(dispatch);
          return <Redirect to={{ pathname: "/login" }} />;
        }

        // authorised so return component
        return <Component {...props} />;
      }}
    />
  );
};
