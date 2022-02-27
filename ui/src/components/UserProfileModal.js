import React, { useState } from "react";
import Form from "react-bootstrap/Form";
import { Button, Card } from "react-bootstrap";
import apiService from "../services/api.service";

import { Alert } from "react-bootstrap";

export default function UserProfileModal(params) {
  const { user, onSave, headerText, onClose } = params;

  const [formErrors, setFormErrors] = useState({});
  const [resetPasswordForm, setResetPasswordForm] = useState({
    newPassword: "",
    email: user?.email,
  });

  function handleChange({ target }) {
    setResetPasswordForm({ ...resetPasswordForm, [target.name]: target.value });
  }

  function formIsValid() {
    const _errors = {};

    // if (!resetPasswordForm.newPassword)
    //   _errors.error = "newPassword is required";
    //
    if (!resetPasswordForm.email) _errors.error = "email is required";

    setFormErrors(_errors);

    return Object.keys(_errors).length === 0;
  }

  async function handleSubmit(event) {
    event.preventDefault();

    if (!formIsValid()) return;

    try {
      await apiService.updateuser({
        userid: user.userid,
        email: resetPasswordForm.email,
        newPassword: resetPasswordForm.newPassword,
      });
      onSave();
    } catch (e) {
      setFormErrors({ error: e.toString() });
    }
  }

  if (!user) return null;
  return (
    <Form onSubmit={handleSubmit}>
      <Card bg="dark" text="white">
        <Card.Header>
          <span>{headerText}</span>
        </Card.Header>
        <Card.Body>
          <div>
            <Alert variant="danger" hidden={!formErrors.error}>
              <Alert.Heading>An Error Occurred</Alert.Heading>
              {formErrors.error}
            </Alert>

            <Form.Label>UserID</Form.Label>
            <Form.Control
              className="font-weight-bold"
              placeholder=""
              value={user.userid}
              disabled
            />
            <Form.Label>Email</Form.Label>
            <Form.Control
              type="email"
              className="font-weight-bold"
              placeholder="Enter email"
              name="email"
              value={resetPasswordForm.email}
              onChange={handleChange}
            />
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
          </div>
        </Card.Body>
        <Card.Footer style={{ display: "flex", flex: "10", gap: "15px" }}>
          <Button variant="primary" type="submit">
            Save
          </Button>
          <Button onClick={onClose}>Close</Button>
        </Card.Footer>
      </Card>
    </Form>
  );
}
