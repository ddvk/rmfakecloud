import React, { useState } from "react";
import Form from "react-bootstrap/Form";
import { Button, Card } from "react-bootstrap";
import apiService from "../../services/api.service";
import { formatDate } from "../../common/date";

import { Alert } from "react-bootstrap";

export default function UserProfileModal(params) {
  const { user, onSave, headerText, onClose } = params;

  const [formErrors, setFormErrors] = useState({});
  const [resetPasswordForm, setResetPasswordForm] = useState({
    newPassword: "",
    email: user?.email,
    quotaGb: user?.quotaBytes ? (Number(user.quotaBytes) / (1024 ** 3)).toFixed(2) : "0",
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
    const quotaGb = Number(resetPasswordForm.quotaGb);
    if (!Number.isFinite(quotaGb) || quotaGb < 0) {
      _errors.error = "Quota must be a number >= 0";
    }

    setFormErrors(_errors);

    return Object.keys(_errors).length === 0;
  }

  async function handleSubmit(event) {
    event.preventDefault();

    if (!formIsValid()) return;

    try {
      const quotaBytes = Math.round(Number(resetPasswordForm.quotaGb) * (1024 ** 3));
      await apiService.updateuser({
        userid: user.userid,
        email: resetPasswordForm.email,
        newPassword: resetPasswordForm.newPassword,
        quotaBytes,
      });
      onSave();
    } catch (e) {
      setFormErrors({ error: e.toString() });
    }
  }

  if (!user) return null;
  return (
    <Form onSubmit={handleSubmit}>
      <Card>
        <Card.Header>
          <span>{headerText}</span>
        </Card.Header>
        <Card.Body>
          <div>
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
            <Form.Group controlId="formPasswordRepeat" className="form-wrapper">
              <Form.Label>New Password</Form.Label>
              <Form.Control
                type="password"
                placeholder="new password"
                value={resetPasswordForm.newPassword}
                name="newPassword"
                onChange={handleChange}
              />
            </Form.Group>
            <Form.Group controlId="formQuotaGb" className="form-wrapper">
              <Form.Label>Quota (GB)</Form.Label>
              <Form.Control
                type="number"
                step="0.01"
                min="0"
                placeholder="0 = Unlimited"
                value={resetPasswordForm.quotaGb}
                name="quotaGb"
                onChange={handleChange}
              />
              <Form.Text muted>Set to 0 for unlimited storage.</Form.Text>
            </Form.Group>
            <Form.Label>Password Changed</Form.Label>
            <Form.Control
              className="font-weight-bold"
              value={user.PasswordChangedAt ? formatDate(user.PasswordChangedAt) : "—"}
              disabled
            />

            <Alert variant="danger" hidden={!formErrors.error}>
              <Alert.Heading>An Error Occurred</Alert.Heading>
              {formErrors.error}
            </Alert>
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
