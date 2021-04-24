let user = localStorage.getItem("currentUser") 
? JSON.parse(localStorage.getItem("currentUser")) 
: null
let token = localStorage.getItem("token")

console.log(token)
console.log(user)
export const initialState = {
  user: "" || user,
  token: "" || token,
  loading: false,
  errorMessage: null,
};
export const AuthReducer = (initialState, action) => {
  switch (action.type) {
    case "REQUEST_LOGIN":
      return {
        ...initialState,
        loading: true,
      };
    case "LOGIN_SUCCESS":
      return {
        ...initialState,
        user: action.payload.user,
        token: action.payload.token,
        loading: false,
      };
    case "LOGOUT":
      return {
        ...initialState,
        user: "",
        token: "",
      };

    case "LOGIN_ERROR":
      return {
        ...initialState,
        loading: false,
        errorMessage: action.error,
      };

    default:
      throw new Error(`Unhandled action type: ${action.type}`);
  }
};
