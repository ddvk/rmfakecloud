import jwt_decode from "jwt-decode";
// import Role from "./Role";
import constants from "../../common/constants"


export async function loginUser(dispatch, loginPayload) {
  const requestOptions = {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(loginPayload),
  };

  try {
    dispatch({ type: "REQUEST_LOGIN" });

    let response = await fetch(`${constants.ROOT_URL}/login`, requestOptions);
    if (!response.ok) {
      dispatch({ type: "LOGIN_ERROR", error: "login failed" });
      return;
    }

    var token = await response.text();
    if (!token) {
      throw new Error("cant retrieve the token");
    }

    var user = jwt_decode(token);

    if (user) {
      dispatch({
        type: "LOGIN_SUCCESS",
        payload: {
          user: user,
          token: token,
        },
      });
      localStorage.setItem("token", token);
      localStorage.setItem("currentUser", JSON.stringify(user));
      return user;
    }

    return;
  } catch (error) {
    console.log(error);
    dispatch({ type: "LOGIN_ERROR", error: "Something went wrong" });
  }
}

export async function logout(dispatch) {
  dispatch({ type: "LOGOUT" });
  localStorage.removeItem("currentUser");
  localStorage.removeItem("token");
}
