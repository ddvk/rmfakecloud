import React, {useState, useLayoutEffect} from "react";
import apiservice from "../services/api.service"
import Stack from 'react-bootstrap/Stack';
import Button from 'react-bootstrap/Button';
import { FaRepeat } from "react-icons/fa6";

export default function CodeGenerator() {

  const [code, setCode] = useState("")
  const [error, setError] = useState("")

  const newCode = async () => {
    setCode("")
    const code = await apiservice.getCode()
      .catch(e => {
        setError(e)
      })
    setCode(code)
  }

  useLayoutEffect(() => {
    newCode()
  }, [])

  if (error) {
    return <div>{error.message}</div>;
  }

  return (
    <>
      <Stack gap={5} style={{alignItems: 'center', marginTop: '15vh'}}>
        <div className="p-2">
          <Button onClick={newCode}>
            <FaRepeat />
          </Button>
        </div>
        <div className="p-2">
          <h1 style={{ letterSpacing: "10px" }}>{code}</h1>
        </div>
      </Stack>
    </>
  );
}
