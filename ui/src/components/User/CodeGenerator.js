import React, {useState} from "react";
import apiservice from "../../services/api.service"


export default function CodeGenerator() {

    const [code, setCode] = useState("")
    const [error, setError] = useState("")

    const newCode = () => {
        setCode("")
        apiservice.getCode()
            .then(x => {
                setCode(x)
            }).catch(e => {
                setError(e)
            })
    }

    if (error) {
        return <div>{error.message}</div>;
    }

    const style = {textAlign:"center", height:"5em"}
    return (
        <>
            <div style={style}> <button onClick={newCode}>Generate Code</button> </div>
            <div style={style}><h1 style={{letterSpacing:"10px"}}>{code}</h1></div>
        </>
    );
}
