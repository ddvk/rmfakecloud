import React, { useState } from "react";
import { useHistory } from "react-router-dom";

import NoMatch from "./NoMatch";
import Form from "react-bootstrap/Form";
import Button from "react-bootstrap/Button";
import Spinner from "./Spinner";
import useFetch from "../hooks/useFetch";
import apiService from "../services/api.service";

import { useParams } from "react-router-dom";

const userListUrl = "users";

export default function UserProfile() {
  const { userid } = useParams();
  const { data: user, loading, error } = useFetch(`${userListUrl}/${userid}`);
  const history = useHistory();

  const [formErrors, setFormErrors] = useState({});
  const [resetPasswordForm, setResetPasswordForm] = useState({
    newPassword: "",
  });

  function handleChange({ target }) {
    setResetPasswordForm({ ...resetPasswordForm, [target.name]: target.value });
  }

  function formIsValid() {
    const _errors = {};

    if (!resetPasswordForm.newPassword)
      _errors.error = "newPassword is required";

    setFormErrors(_errors);

    return Object.keys(_errors).length === 0;
  }

  async function handleSubmit(event) {
    event.preventDefault();

    if (!formIsValid()) return;

    try {
      await apiService.updateuser({
        userid,
        newPassword: resetPasswordForm.newPassword,
      });
      history.push("/userList")
      
    } catch (e) {
      setFormErrors({ error: e.toString()});
    }
  }

  if (loading) return <Spinner />;
  if (error) {
    return <div>{error.message}</div>;
  }
  if (!user) return <NoMatch />;

  return (
    <div>
      <Form onSubmit={handleSubmit}>
        <Form.Group controlId="formEmail">
          <Form.Label>Email address</Form.Label>
          <Form.Control
            type="email"
            className="font-weight-bold"
            placeholder="Enter email"
            value={userid}
            disabled
          />
        </Form.Group>
        <Form.Group controlId="formPasswordRepeat">
          <Form.Label>New Password</Form.Label>
          <Form.Control
            type="password"
            placeholder="new password"
            value={resetPasswordForm.newPassword}
            name="newPassword"
            onChange={handleChange}
          />
        </Form.Group>
        {formErrors.error && (
          <div className="alert alert-danger">{formErrors.error}</div>
        )}
        <Button variant="primary" type="submit">
          Save
        </Button>
      </Form>
    </div>
  );
}
