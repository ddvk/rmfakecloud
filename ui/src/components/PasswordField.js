import React, { useState, useRef, useEffect } from "react";
import { Form, InputGroup, Button } from "react-bootstrap";
import { FaEye } from "react-icons/fa";

function PasswordField(props) {
  const [inputType, setInputType] = useState("password");
  const [inputCursorPosition, setInputCursorPosition] = useState(0);
  const inputEl = useRef(null);

  function showHide(e) {
    e.preventDefault();
    e.stopPropagation();

    setInputType(inputType === "text" ? "password" : "text");

    inputEl.current.focus();
  }

  useEffect(() => {
    const init = () => {
      inputEl.current.selectionStart = inputCursorPosition;
    };
    init();
  }, [inputType, inputCursorPosition]);

  function saveCursorPosition(e) {
    setInputCursorPosition(e.target.selectionStart);
  }

  return (
    <InputGroup className="mb-3">
      <Form.Control
        type={inputType}
        {...props}
        ref={inputEl}
        onBlur={saveCursorPosition}
      />
      <InputGroup.Append>
        <Button onClick={showHide}>
          <FaEye />
        </Button>
      </InputGroup.Append>
    </InputGroup>
  );
}

export default PasswordField;
