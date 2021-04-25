import React, { useState } from "react";
import apiservice from "../../services/api.service";
import Form from "react-bootstrap/Form";
import Button from "react-bootstrap/Button";
import { useAuthState, useAuthDispatch } from "../../hooks/useAuthContext";
import PasswordField from "../PasswordField";
import { logout } from "../Login/actions";

export default function OwnUserProfile() {
  const { user: loggedInUser } = useAuthState();
  const authDispatch = useAuthDispatch();

  const [formErrors, setFormErrors] = useState({});
  const [resetPasswordForm, setResetPasswordForm] = useState({
    email: loggedInUser.Email,
    currentPassword: null,
    newPassword: null,
    confirmNewPassword: null,
  });

  function handleChange({ target }) {
    setResetPasswordForm({ ...resetPasswordForm, [target.name]: target.value });
  }

  function formIsValid() {
    const _errors = {};

    if (!resetPasswordForm.email) _errors.email = "email is required";
    if (!resetPasswordForm.currentPassword)
      _errors.currentPassword = "currentPassword id is required";
    if (!resetPasswordForm.newPassword)
      _errors.newPassword = "newPassword is required";
    if (!resetPasswordForm.confirmNewPassword)
      _errors.confirmNewPassword = "confirm new password is required";
    if (resetPasswordForm.confirmNewPassword !== resetPasswordForm.newPassword)
      _errors.confirmNewPassword =
        "confirm new password must match new password";

    setFormErrors(_errors);

    return Object.keys(_errors).length === 0;
  }

  function handleSubmit(event) {
    event.preventDefault();

    if (!formIsValid()) return;

    apiservice.resetPassword(resetPasswordForm)
    .then(r => {
      if (r.ok) {
        logout(authDispatch)
        return
      }
      if (r.status === 400) {
        return r.json()
      }
      throw new Error("unknown error: " + r.status)
    })
    .then(j => {
      if (j && j.error){
        setFormErrors(j)
        return
      } else {
        setFormErrors({error:"invalid data"})
      }
    })
    .catch(e => {
      console.log(e)
      setFormErrors({
        error:e.toString()
      })
    })
    
  }

  return (
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
          name="currentPassword"
          type="password"
          placeholder="current password"
          value={resetPasswordForm.currentPassword}
          onChange={handleChange}
        />
        {formErrors.currentPassword && (
          <div className="alert alert-danger">{formErrors.currentPassword}</div>
        )}
      </Form.Group>
      <Form.Group controlId="formNewPassword">
        <Form.Label>New Password</Form.Label>
        <PasswordField
          name="newPassword"
          placeholder="new password"
          value={resetPasswordForm.newPassword}
          onChange={handleChange}
        />
        {formErrors.newPassword && (
          <div className="alert alert-danger">{formErrors.newPassword}</div>
        )}
      </Form.Group>
      <Form.Group controlId="formConfirmNewPassword">
        <Form.Label>Confirm Password</Form.Label>
        <PasswordField
          name="confirmNewPassword"
          placeholder="confirm new password"
          value={resetPasswordForm.confirmNewPassword}
          onChange={handleChange}
        />
        {formErrors.confirmNewPassword && (
          <div className="alert alert-danger">
            {formErrors.confirmNewPassword}
          </div>
        )}
      </Form.Group>
        {formErrors.error && (
          <div className="alert alert-danger">
            {formErrors.error}
          </div>
        )}

      <Button variant="primary" type="submit">
        Save
      </Button>
    </Form>
  );
}
