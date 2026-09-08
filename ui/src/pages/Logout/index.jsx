import { useEffect } from "react";
import { Redirect } from "react-router-dom";
import Spinner from "../../components/Spinner";
import { logout } from "../../common/actions";
import { useAuthState } from "../../common/useAuthContext";

export default function LogoutPage() {
  const { dispatch } = useAuthState();

  useEffect(() => {
    logout(dispatch);
  }, [dispatch]);

  return (
    <>
      <Spinner />
      <Redirect to="/login" />
    </>
  );
}
