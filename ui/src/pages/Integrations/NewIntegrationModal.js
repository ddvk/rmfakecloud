import React, { useState } from "react";
import Form from "react-bootstrap/Form";
import { Button, Card } from "react-bootstrap";
import apiService from "../../services/api.service";

import { Alert } from "react-bootstrap";

export default function IntegrationProfileModal(params) {
  const { onSave, onClose } = params;

  const [formErrors, setFormErrors] = useState({});
  const [formInfo, setFormInfo] = useState({});
  const [integrationForm, setIntegrationForm] = useState({
    name: "",
    provider: "localfs",
  });

  function handleChange({ target }) {
    setIntegrationForm({ ...integrationForm, [target.name]: target.value });
  }

  function formIsValid() {
    const _errors = {};

    if (!integrationForm.name) _errors.error = "name is required";

    if (!integrationForm.provider) _errors.error = "provider is required";

    setFormErrors(_errors);

    return Object.keys(_errors).length === 0;
  }

  async function handleSubmit(event) {
    event.preventDefault();

    if (!formIsValid()) return;

    try {
      await apiService.createintegration(integrationForm);
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
          <span>New Integration</span>
        </Card.Header>
        <Card.Body>
          <Alert variant="danger" hidden={!formErrors.error}>
            <Alert.Heading>An Error Occurred</Alert.Heading>
            {formErrors.error}
          </Alert>

          <Alert variant="info" hidden={!formInfo.message}>
            {formInfo.message}
          </Alert>

          <Form.Label>Name</Form.Label>
          <Form.Control
            placeholder="Integration name"
            value={integrationForm.name}
            name="name"
            autofocus
            onChange={handleChange}
          />

          <Form.Label>Provider</Form.Label>
          <Form.Select
            name="provider"
            value={integrationForm.provider}
            onChange={handleChange}
            className="mb-1"
          >
            <option value="localfs">Directory in file system</option>
            <option value="webdav">WebDAV</option>
            <option value="dropbox">Dropbox</option>
          </Form.Select>

          {integrationForm.provider === "webdav" && (
            <>
              <Form.Label>Address</Form.Label>
              <Form.Control
                placeholder="Server URL"
                value={integrationForm.address}
                name="address"
                onChange={handleChange}
              />
            </>
          )}
          {integrationForm.provider === "webdav" && (
            <>
              <Form.Label>Username</Form.Label>
              <Form.Control
                placeholder="Username"
                value={integrationForm.username}
                name="username"
                onChange={handleChange}
              />
            </>
          )}
          {integrationForm.provider === "webdav" && (
            <>
              <Form.Label>Password</Form.Label>
              <Form.Control
                type="password"
                placeholder="Password"
                value={integrationForm.password}
                name="password"
                onChange={handleChange}
              />
            </>
          )}

          {integrationForm.provider === "localfs" && (
            <>
              <Form.Label>Path</Form.Label>
              <Form.Control
                placeholder="Path"
                value={integrationForm.path}
                name="path"
                onChange={handleChange}
              />
            </>
          )}

          {integrationForm.provider === "dropbox" && (
            <>
              <Form.Label>Access Token</Form.Label>
              <Form.Control
                placeholder="Access Token"
                value={integrationForm.accesstoken}
                name="accesstoken"
                onChange={handleChange}
              />
            </>
          )}
        </Card.Body>
        <Card.Footer style={{ display: "flex", gap: "15px" }}>
          <Button variant="primary" type="submit">
            Save
          </Button>
          <Button variant="secondary" onClick={onClose}>Close</Button>
        </Card.Footer>
      </Card>
    </Form>
  );
}
