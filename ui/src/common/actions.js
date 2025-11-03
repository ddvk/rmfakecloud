import apiService from "../services/api.service";

export async function loginUser(dispatch, loginPayload) {

  try {
    dispatch({ type: "REQUEST_LOGIN" });

    let user = await apiService.login(loginPayload)
    dispatch({
      type: "LOGIN_SUCCESS",
      payload: { user: user },
    });

    return;
  } catch (error) {
    dispatch({ type: "LOGIN_ERROR", error: "Can't login: " + error.message});
  }
}

export async function logout(dispatch) {
  await apiService.logout()
  dispatch({ type: "LOGOUT" });
}

export async function oidcAuth(dispatch) {

  try {
    dispatch({ type: "REQUEST_LOGIN" });

    return await apiService.oidcAuth();
  } catch (error) {
    dispatch({ type: "LOGIN_ERROR", error: "Can't login: " + error.message});
  }
}

export async function oidcCallback(dispatch, callbackPayload) {
  try {
    dispatch({ type: "REQUEST_LOGIN" });

    let user = await apiService.oidcCallback(callbackPayload)
    dispatch({
      type: "LOGIN_SUCCESS",
      payload: { user: user },
    });

    return;
  } catch (error) {
    dispatch({ type: "LOGIN_ERROR", error: "Can't login: " + error.message});
  }
}