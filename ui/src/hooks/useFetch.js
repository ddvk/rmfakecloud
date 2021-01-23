import { useState, useEffect } from "react";

const baseUrl = process.env.REACT_APP_API_BASE_URL;

const useFetch = (url) => {
  const [loading, setLoading] = useState(true);
  const [data, setData] = useState(null);
  const [error, setError] = useState(null);

  useEffect(() => {
    const init = async () => {
      try {
        //const response = await fetch(baseUrl + url, {
        console.log(baseUrl);

        debugger;
        const response = await fetch(url, {
          method: "GET",
          headers: new Headers({
            Authorization: "some_token",
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
