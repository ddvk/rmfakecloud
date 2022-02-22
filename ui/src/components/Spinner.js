import React from "react";
import { Card, Spinner as BSpinner } from "react-bootstrap";

export default function Spinner() {
  return (
    <Card
      bg="dark"
      text="white"

      style={{
        padding: "8px",
        display: "flex",
        width: "fit-content",
        justifyContent: "space-between",
        flexDirection: "row",
        gap: "8px",
        lineHeight: "1.75em",
        margin: "auto",
      }}
    >
      <BSpinner animation="grow" role="status" />
      <span>Loading...</span>
    </Card>
  );
}
