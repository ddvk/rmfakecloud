import React, {useEffect} from "react";
import {oidcCallback} from "../../common/actions.js";
import apiService from "../../services/api.service.js";

const Logout = () => {
  useEffect(() => {
    const init = async () => {
      try {
        const json = await apiService.oidcInfo()

        if (json != null && !json.only) {
          window.location.replace("/login")
          return
        }
      } catch (error) {
        console.error(error)
      }
    }

    init()
  }, []);

  return <div>You are now logged out.</div>;
};

export default Logout;
