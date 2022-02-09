import React, { useState, useEffect } from "react";
import { useHistory } from "react-router-dom";

import NoMatch from "./NoMatch";
import Form from "react-bootstrap/Form";
import Button from "react-bootstrap/Button";
import Spinner from "./Spinner";
import useFetch from "../hooks/useFetch";
import apiService from "../services/api.service";
import { toast } from 'react-toastify';
import { useParams } from "react-router-dom";

const userListUrl = "users";

export default function UserProfile() {
  const { userid } = useParams();
  
  const { data: user, loading, error } = useFetch(`${userListUrl}/${userid}`);
  const history = useHistory();

  const [formErrors, setFormErrors] = useState({});
  const [profileForm, setProfileForm] = useState({
    newPassword: "",
    email: "",
  });

  useEffect(() => {
    if (!user)
      return

    setProfileForm(oldState => ({
      ...oldState,
      email: user.email
    }))
  }, [user]);

  function handleChange({ target }) {
    setProfileForm({ ...profileForm, [target.name]: target.value });
  }

  function formIsValid() {
    const _errors = {};

    // if (!profileForm.newPassword)
    //   _errors.error = "newPassword is required";

    setFormErrors(_errors);

    return Object.keys(_errors).length === 0;
  }

  async function handleSubmit(event) {
    event.preventDefault();

    if (!formIsValid()) return;

    try {
      await apiService.updateuser({
        userid,
        newPassword: profileForm.newPassword,
        email: profileForm.email
      });
      toast("updated")
      history.push("/users")
      
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
          <Form.Label>UserId</Form.Label>
          <Form.Control
            className="font-weight-bold"
            placeholder={userid}
            value={user.userid}
            disabled
          />
        <Form.Group controlId="formEmail">
          <Form.Label>Email address</Form.Label>
          <Form.Control
            type="email"
            className="font-weight-bold"
            placeholder="Enter email"
            name="email"
            value={profileForm.email}
            onChange={handleChange}
          />
        </Form.Group>
        <Form.Group controlId="formPasswordRepeat">
          <Form.Label>New Password</Form.Label>
          <Form.Control
            type="password"
            placeholder="new password"
            value={profileForm.newPassword}
            name="newPassword"
            onChange={handleChange}
          />
        </Form.Group>
        <div>
          <p>Integrations</p>
          <ul>
          {user.integrations && user.integrations.map((x, i) => 
            (<li>{x}</li>)
          )}
          </ul>
        </div>
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
