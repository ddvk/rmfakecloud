import React from "react";

import NoMatch from "../NoMatch";
import Form from "react-bootstrap/Form";
import Button from "react-bootstrap/Button";
import Spinner from "../Spinner";
import useFetch from "../../hooks/useFetch";

import { useParams } from "react-router-dom";

const userListUrl = "users";

export default function UserProfile() {
  debugger;

  const { userid } = useParams();
  //const navigate = useNavigate();

  const { data: user, loading, error } = useFetch(`${userListUrl}/${userid}`);

  if (loading) return <Spinner />;
  if (error) {
    return <div>{error.message}</div>;
  }
  if (!user) return <NoMatch />;

  return (
    <div>
      <Form>
        <Form.Group controlId="formEmail">
          <Form.Label>Email address</Form.Label>
          <Form.Control
            type="email"
            className="font-weight-bold"
            placeholder="Enter email"
            value={user.email}
            disabled
          />
        </Form.Group>
        <Form.Group controlId="formPassword">
          <Form.Label>One Time Password</Form.Label>
          <Form.Control type="password" placeholder="password" />
        </Form.Group>
        <Form.Group controlId="formPasswordRepeat">
          <Form.Label>Repeat One Time Password</Form.Label>
          <Form.Control
            type="password"
            placeholder="repeat one time password"
          />
        </Form.Group>
        <Button variant="primary" type="submit">
          Save
        </Button>
      </Form>
    </div>
  );
}
