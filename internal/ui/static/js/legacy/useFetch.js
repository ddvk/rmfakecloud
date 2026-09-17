(function (global) {
  "use strict";
  var RM = global.RM;
  function useFetch(url, options) {
    var loadingState = RM.useState(true);
    var loading = loadingState[0];
    var setLoading = loadingState[1];
    var dataState = RM.useState(null);
    var data = dataState[0];
    var setData = dataState[1];
    var errState = RM.useState(null);
    var error = errState[0];
    var setError = errState[1];
    var constants = global.constants || { ROOT_URL: "/ui/api" };

    RM.useEffect(
      function () {
        var cancelled = false;
        (async function () {
          try {
            var response = await fetch(constants.ROOT_URL + "/" + url, {
              method: "GET",
              credentials: "same-origin",
            });
            if (cancelled) return;
            if (response.ok) {
              setData(await response.json());
            } else if (response.status === 401) {
              localStorage.removeItem("currentUser");
              window.location.replace("/");
            } else {
              throw response;
            }
          } catch (e) {
            if (!cancelled) {
              console.error("fetch failed: ", e);
              setError(e);
            }
          } finally {
            if (!cancelled) setLoading(false);
          }
        })();
        return function () {
          cancelled = true;
        };
      },
      [url, options]
    );

    return { data: data, error: error, loading: loading };
  }
  global.useFetch = useFetch;
})(typeof window !== "undefined" ? window : globalThis);
