import React, { useState } from "react";

import NoMatch from "../NoMatch";
import Form from "react-bootstrap/Form";
import Button from "react-bootstrap/Button";
import Spinner from "../Spinner";
import useFetch from "../../hooks/useFetch";

import { useParams } from "react-router-dom";

const userListUrl = "users";

export default function UserProfile() {
  const { userid } = useParams();
  const { data: user, loading, error } = useFetch(`${userListUrl}/${userid}`);

  const [_, setFormErrors] = useState({});
  const [resetPasswordForm, setResetPasswordForm] = useState({
    email: user.email,
    currentPassword: null,
    newPassword: null,
  });

  function handleChange({ target }) {
    setResetPasswordForm({ ...resetPasswordForm, [target.name]: target.value });
  }

  function formIsValid() {
    const _errors = {};

    if (!resetPasswordForm.title) _errors.email = "email is required";
    if (!resetPasswordForm.authorId)
      _errors.currentPassword = "currentPassword id is required";
    if (!resetPasswordForm.category)
      _errors.newPassword = "newPassword is required";

    setFormErrors(_errors);

    return Object.keys(_errors).length === 0;
  }

  function handleSubmit(event) {
    event.preventDefault();

    if (!formIsValid()) return;

    console.log("saving user profile.");

    // courseApi.saveCourse(course).then(() => {
    //   props.history.push("/courses");
    //   console.log("calling toast");
    //   toast.success("Bravo");
    // });
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
            value={resetPasswordForm.email}
            disabled
          />
        </Form.Group>
        <Form.Group controlId="formPassword">
          <Form.Label>Old Password</Form.Label>
          <Form.Control
            type="password"
            placeholder="current password"
            value={resetPasswordForm.currentPassword}
            onChange={handleChange}
          />
        </Form.Group>
        <Form.Group controlId="formPasswordRepeat">
          <Form.Label>New Password</Form.Label>
          <Form.Control
            type="password"
            placeholder="new password"
            value={resetPasswordForm.newPassword}
            onChange={handleChange}
          />
        </Form.Group>
        <Button variant="primary" type="submit">
          Save
        </Button>
      </Form>
    </div>
  );
}
