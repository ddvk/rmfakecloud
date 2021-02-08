import { useState, useEffect, useContext } from "react";
import { useAuthState } from "./useAuthContext";
const ROOT_URL = `${window.location.origin}/ui/api`;

const useFetch = (url) => {
  const [loading, setLoading] = useState(true);
  const [data, setData] = useState(null);
  const [error, setError] = useState(null);
  const { token } = useAuthState();

  useEffect(() => {
    const init = async () => {
      try {
        console.log(`${ROOT_URL}/${url}`);
        const response = await fetch(`${ROOT_URL}/${url}`, {
          method: "GET",
          headers: new Headers({
            Authorization: `Bearer ${token}`,
          }),
        });

        if (response.ok) {
          const json = await response.json();
          setData(json);
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
  }, [url]); // rerun when...

  return { data, error, loading };
};

export default useFetch;
