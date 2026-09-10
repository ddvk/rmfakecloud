(function (global) {
  "use strict";
  var RM = global.RM;
  if (!RM) throw new Error("rm.js must load before auth.js");

  var user = null;
  try {
    user = localStorage.getItem("currentUser")
      ? JSON.parse(localStorage.getItem("currentUser"))
      : null;
  } catch (_) {
    user = null;
  }

  var store = {
    state: {
      user: user || "",
      loading: false,
      errorMessage: null,
    },
    listeners: [],
  };

  function emit() {
    store.listeners.slice().forEach(function (fn) {
      try {
        fn();
      } catch (e) {
        console.error(e);
      }
    });
  }

  function authReducer(state, action) {
    switch (action.type) {
      case "REQUEST_LOGIN":
        return Object.assign({}, state, { loading: true });
      case "LOGIN_SUCCESS":
        return Object.assign({}, state, { user: action.payload.user, loading: false, errorMessage: null });
      case "LOGOUT":
        return Object.assign({}, state, { user: "" });
      case "LOGIN_ERROR":
        return Object.assign({}, state, { loading: false, errorMessage: action.error });
      default:
        throw new Error("Unhandled action type: " + action.type);
    }
  }

  function dispatch(action) {
    store.state = authReducer(store.state, action);
    emit();
  }

  function useAuthState() {
    var tick = RM.useState(0);
    var setTick = tick[1];
    RM.useEffect(function () {
      var sub = function () {
        setTick(function (n) {
          return n + 1;
        });
      };
      store.listeners.push(sub);
      return function () {
        store.listeners = store.listeners.filter(function (x) {
          return x !== sub;
        });
      };
    }, []);
    return { state: store.state, dispatch: dispatch };
  }

  function AuthProvider(props) {
    return RM.h("div", { className: "rm-auth-provider", style: { display: "contents" } }, props.children);
  }

  async function loginUser(dispatchFn, loginPayload) {
    var d = dispatchFn || dispatch;
    try {
      d({ type: "REQUEST_LOGIN" });
      var u = await global.apiService.login(loginPayload);
      d({ type: "LOGIN_SUCCESS", payload: { user: u } });
    } catch (error) {
      d({ type: "LOGIN_ERROR", error: "Can't login: " + error.message });
    }
  }

  async function logout(dispatchFn) {
    var d = dispatchFn || dispatch;
    await global.apiService.logout();
    d({ type: "LOGOUT" });
  }

  global.useAuthState = useAuthState;
  global.AuthProvider = AuthProvider;
  global.loginUser = loginUser;
  global.logoutAction = logout;
})(typeof window !== "undefined" ? window : globalThis);
