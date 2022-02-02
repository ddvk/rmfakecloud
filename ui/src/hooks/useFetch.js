import { useState, useEffect } from "react";
import { useAuthState} from "../common/useAuthContext";
import constants from "../common/constants"


const useFetch = (url, options) => {
  const [loading, setLoading] = useState(true);
  const [data, setData] = useState(null);
  const [error, setError] = useState(null);
  const { dispatch } = useAuthState();

  useEffect(() => {
    const init = async () => {
      try {
        const response = await fetch(`${constants.ROOT_URL}/${url}`, {
          method: "GET",
        });

        if (response.ok) {
          const json = await response.json();
          setData(json);
        } else if (response.status === 401) {
          //logout(dispatch);
          // //TODO: fix this hack
          localStorage.removeItem("currentUser");
          window.location.replace("/")
        } else {
          throw response;
        }
      } catch (e) {
        console.error("fetch failed: ", e);
        setError(e);
      } finally {
        setLoading(false);
      }
    };

    init();
  }, [url, dispatch]); // rerun when...

  return { data, error, loading };
};
export default useFetch
