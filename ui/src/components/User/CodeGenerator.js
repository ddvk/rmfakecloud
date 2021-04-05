import React from "react";

import Spinner from "../Spinner";
import useFetch from "../../hooks/useFetch";

const codeUrl = "newcode";

export default function CodeGenerator() {
  const { data: code, loading, error } = useFetch(`${codeUrl}`);

  if (loading) return <Spinner />;
  if (error) {
    return <div>{error.message}</div>;
  }

  return (
    <>
      <div style={{ margin: "auto", width: "30%" }}>
        <p>Onetime Code</p>
        <h1>{code}</h1>
      </div>
    </>
  );
}
