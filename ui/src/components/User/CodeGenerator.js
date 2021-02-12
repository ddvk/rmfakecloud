import React from "react";

import NoMatch from "../NoMatch";
import Form from "react-bootstrap/Form";
import Button from "react-bootstrap/Button";
import Spinner from "../Spinner";
import useFetch from "../../hooks/useFetch";

import { useParams } from "react-router-dom";

const userListUrl = "newcode";

export default function CodeGenerator() {
  const { data: code, loading, error } = useFetch(`${userListUrl}`);

  if (loading) return <Spinner />;
  if (error) {
    return <div>{error.message}</div>;
  }

  return (
    <>

    <div style={{"margin":"auto", "width":"30%"}}>
      <p>Onetime Code</p>
        <h1>{code}</h1>
    </div>
    </>
      
  );
}
