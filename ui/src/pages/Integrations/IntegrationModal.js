import React, { useState } from "react";
import Form from "react-bootstrap/Form";
import { Button, Card } from "react-bootstrap";
import apiService from "../../services/api.service";

import { Alert } from "react-bootstrap";

export default function IntegrationModal(params) {
  const { integration, onSave, headerText, onClose } = params;

  const [formErrors, setFormErrors] = useState({});
  const [integrationForm, setIntegrationForm] = useState({
    name: integration?.Name,
    provider: integration?.Provider,
    email: integration?.email,
    username: integration?.Username,
    password: integration?.Password,
    address: integration?.Address,
    activetransfers: integration?.ActiveTransfers,
    insecure: integration?.Insecure,
    accesstoken: integration?.Accesstoken,
    path: integration?.Path,
  });

  function handleChange({ target }) {
    setIntegrationForm({ ...integrationForm, [target.name]: target.value });
  }

  function formIsValid() {
    const _errors = {};

    if (!integrationForm.name) _errors.error = "name is required";

    setFormErrors(_errors);

    return Object.keys(_errors).length === 0;
  }

  async function handleSubmit(event) {
    event.preventDefault();

    if (!formIsValid()) return;

    try {
      await apiService.updateintegration({
        id: integration.ID,
        name: integrationForm.name,
        provider: integrationForm.provider,
        username: integrationForm.username,
        password: integrationForm.password,
        address: integrationForm.address,
        activetransfers: integrationForm?.activetransfers,
        insecure: integrationForm.insecure,
        accesstoken: integrationForm.accesstoken,
        path: integrationForm.path,
      });
      onSave();
    } catch (e) {
      setFormErrors({ error: e.toString() });
    }
  }

  if (!integration) return null;
  return (
    <Form onSubmit={handleSubmit}>
      <Card>
        <Card.Header>
          <span>{headerText}</span>
        </Card.Header>
        <Card.Body>
          <div>
            <Alert variant="danger" hidden={!formErrors.error}>
              <Alert.Heading>An Error Occurred</Alert.Heading>
              {formErrors.error}
            </Alert>

            <Form.Label>IntegrationID</Form.Label>
            <Form.Control
              className="font-weight-bold"
              placeholder=""
              value={integration.ID}
              disabled
            />

            <Form.Label>Provider</Form.Label>
            <Form.Control
              as="select"
              name="provider"
              value={integrationForm.provider}
              onChange={handleChange}
            >
              <option value="localfs">Directory in file system</option>
              <option value="webdav">WebDAV</option>
              <option value="ftp">FTP</option>
              <option value="dropbox">Dropbox</option>
            </Form.Control>

            <Form.Label>Name</Form.Label>
            <Form.Control
              placeholder="Integration name"
              value={integrationForm.name}
              name="name"
              onChange={handleChange}
            />

            {(integrationForm.provider === "webdav" || integrationForm.provider === "ftp") && (
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
            {(integrationForm.provider === "webdav" || integrationForm.provider === "ftp") && (
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
            {(integrationForm.provider === "webdav" || integrationForm.provider === "ftp") && (
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

            {integrationForm.provider === "ftp" && (
              <Form.Check
                name="activetransfers"
                checked={integrationForm.activetransfers}
                onChange={({ target }) => setIntegrationForm({ ...integrationForm, [target.name]: target.checked })}
                label="Use actives transfers"
              />
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
          </div>
        </Card.Body>
        <Card.Footer style={{ display: "flex", flex: "10", gap: "15px" }}>
          <Button variant="primary" type="submit">
            Save
          </Button>
          <Button variant="secondary" onClick={onClose}>Cancel</Button>
        </Card.Footer>
      </Card>
    </Form>
  );
}
