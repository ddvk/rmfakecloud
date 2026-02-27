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
    headers: [],
  });

  function handleChange({ target }) {
    setIntegrationForm({ ...integrationForm, [target.name]: target.value });
  }

  function handleHeaderChange(index, field, value) {
    const newHeaders = [...integrationForm.headers];
    newHeaders[index] = { ...newHeaders[index], [field]: value };
    setIntegrationForm({ ...integrationForm, headers: newHeaders });
  }

  function addHeader() {
    setIntegrationForm({
      ...integrationForm,
      headers: [...integrationForm.headers, { name: "", value: "" }],
    });
  }

  function removeHeader(index) {
    const newHeaders = integrationForm.headers.filter((_, i) => i !== index);
    setIntegrationForm({ ...integrationForm, headers: newHeaders });
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

    console.log(integrationForm)
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
            <div style={{'white-space': 'pre-wrap'}}>
              {formErrors.error}
            </div>
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
            <option value="ftp">FTP</option>
            <option value="webdav">WebDAV</option>
            <option value="dropbox">Dropbox</option>
            <option value="webhook">Messaging webhook</option>
            <option value="opds">OPDS Catalog</option>
          </Form.Select>

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

          {integrationForm.provider === "webhook" && (
            <>
              <Form.Label>Endpoint</Form.Label>
              <Form.Control
                placeholder="https://automation.domain.tld/webhook/0123-456789-abc"
                value={integrationForm.endpoint}
                name="endpoint"
                onChange={handleChange}
              />
            </>
          )}

          {integrationForm.provider === "opds" && (
            <>
              <Form.Label>Feed URL</Form.Label>
              <Form.Control
                placeholder="https://example.com/opds"
                value={integrationForm.feedurl}
                name="feedurl"
                onChange={handleChange}
              />

              <Form.Label className="mt-2">HTTP Headers</Form.Label>
              {integrationForm.headers.map((header, index) => (
                <div key={index} style={{ display: "flex", gap: "10px", marginBottom: "5px" }}>
                  <Form.Control
                    placeholder="Header name"
                    value={header.name}
                    onChange={(e) => handleHeaderChange(index, "name", e.target.value)}
                    style={{ flex: 1 }}
                  />
                  <Form.Control
                    placeholder="Header value"
                    value={header.value}
                    onChange={(e) => handleHeaderChange(index, "value", e.target.value)}
                    style={{ flex: 1 }}
                  />
                  <Button variant="outline-danger" onClick={() => removeHeader(index)}>
                    Remove
                  </Button>
                </div>
              ))}
              <div className="mt-2">
                <Button variant="outline-secondary" size="sm" onClick={addHeader}>
                  Add Header
                </Button>
              </div>
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
