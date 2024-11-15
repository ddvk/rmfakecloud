import React, { useState } from "react";
import Form from "react-bootstrap/Form";
import { Button, Card, Alert } from "react-bootstrap";

import apiService from "../../services/api.service";

export default function UserProfileModal(params) {
  const { onSave, onClose } = params;

  const [formErrors, setFormErrors] = useState({});
  const [formInfo, setFormInfo] = useState({});
  const [userForm, setUserForm] = useState({
    newPassword: "",
    email: "",
    userid: "",
  });

  function handleChange({ target }) {
    setUserForm({ ...userForm, [target.name]: target.value });
  }

  function formIsValid() {
    const _errors = {};

    if (!userForm.newPassword) _errors.error = "password is required";

    if (!userForm.email) _errors.error = "email is required";

    if (!userForm.userid) _errors.error = "userid is required";

    setFormErrors(_errors);

    return Object.keys(_errors).length === 0;
  }

  async function handleSubmit(event) {
    event.preventDefault();

    if (!formIsValid()) return;

    try {
      await apiService.createuser(userForm);
      setFormInfo({ message: "Created" });
      onSave();
    } catch (e) {
      setFormErrors({ error: e.toString() });
    }
  }

  return (
    <Form onSubmit={handleSubmit} autoComplete="off">
      <Card>
        <Card.Header>
          <span>New User</span>
        </Card.Header>
        <Card.Body>
          <Alert variant="danger" hidden={!formErrors.error}>
            <Alert.Heading>An Error Occurred</Alert.Heading>
            {formErrors.error}
          </Alert>

          <Alert variant="info" hidden={!formInfo.message}>
            {formInfo.message}
          </Alert>

          <Form.Label>UserID</Form.Label>
          <Form.Control
            autoComplete="blah"
            className="font-weight-bold"
            placeholder=""
            name="userid"
            value={userForm.userid}
            onChange={handleChange}
          />
          <Form.Label>Email</Form.Label>
          <Form.Control
            autoComplete="blah"
            type="email"
            className="font-weight-bold"
            placeholder="Enter email"
            name="email"
            value={userForm.email}
            onChange={handleChange}
          />

          <Form.Group controlId="formPasswordRepeat">
            <Form.Label>Password</Form.Label>
            <Form.Control
              autoComplete="blah"
              type="password"
              placeholder="newPassword"
              value={userForm.newPassword}
              name="newPassword"
              onChange={handleChange}
            />
          </Form.Group>
        </Card.Body>
        <Card.Footer style={{ display: "flex", gap: "15px" }}>
          <Button variant="primary" type="submit">
            Save
          </Button>
          <Button onClick={onClose}>Close</Button>
        </Card.Footer>
      </Card>
    </Form>
  );
}
