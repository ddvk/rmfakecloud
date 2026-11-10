import { useCallback, useEffect, useState } from "react";
import Table from "react-bootstrap/Table";
import Spinner from "react-bootstrap/Spinner";
import Button from "react-bootstrap/Button";
import Form from "react-bootstrap/Form";
import { toast } from "react-toastify";
import { startRegistration, browserSupportsWebAuthn } from "@simplewebauthn/browser";

import apiservice from "../../services/api.service";

function formatWhen(iso) {
  if (!iso) return "—";
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
}

export default function Passkeys() {
  const [enabled, setEnabled] = useState(false);
  const [credentials, setCredentials] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [busy, setBusy] = useState(false);
  const [name, setName] = useState("");

  const load = useCallback(() => {
    setLoading(true);
    return apiservice
      .webAuthnStatus()
      .then((st) => {
        const on = Boolean(st && st.enabled) && browserSupportsWebAuthn();
        setEnabled(on);
        if (!on) {
          setCredentials([]);
          setError(null);
          return null;
        }
        return apiservice.listWebAuthnCredentials();
      })
      .then((data) => {
        if (data == null) return;
        setCredentials(Array.isArray(data) ? data : []);
        setError(null);
      })
      .catch((e) => {
        setError(e.message || String(e));
        setCredentials([]);
      })
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  async function handleAdd() {
    setBusy(true);
    try {
      const begin = await apiservice.webAuthnRegisterBegin();
      const credential = await startRegistration({ optionsJSON: begin.publicKey });
      await apiservice.webAuthnRegisterFinish(begin.sessionId, credential, name.trim() || "Passkey");
      setName("");
      toast.success("Passkey registered.");
      await load();
    } catch (e) {
      toast.error(e.message || String(e));
    } finally {
      setBusy(false);
    }
  }

  async function handleDelete(id) {
    setBusy(true);
    try {
      await apiservice.deleteWebAuthnCredential(id);
      toast.success("Passkey removed.");
      await load();
    } catch (e) {
      toast.error(e.message || String(e));
    } finally {
      setBusy(false);
    }
  }

  if (!enabled && !loading) {
    return null;
  }

  return (
    <section className="mt-4">
      <h2>Passkeys</h2>
      <p className="text-muted">
        Register a passkey for passwordless sign-in on this site. Password login remains available.
      </p>
      {loading ? (
        <Spinner animation="border" role="status" size="sm" />
      ) : error ? (
        <p className="text-danger">{error}</p>
      ) : (
        <>
          <Form
            className="mb-3 d-flex flex-wrap gap-2 align-items-end"
            onSubmit={(e) => {
              e.preventDefault();
              handleAdd();
            }}
          >
            <Form.Group controlId="passkeyName">
              <Form.Label>Label</Form.Label>
              <Form.Control
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. Laptop"
                disabled={busy}
              />
            </Form.Group>
            <Button type="submit" disabled={busy}>
              Add passkey
            </Button>
          </Form>
          {credentials.length === 0 ? (
            <p>No passkeys registered yet.</p>
          ) : (
            <Table striped bordered hover responsive size="sm">
              <thead>
                <tr>
                  <th scope="col">Name</th>
                  <th scope="col">Created</th>
                  <th scope="col">Actions</th>
                </tr>
              </thead>
              <tbody>
                {credentials.map((c) => (
                  <tr key={c.id}>
                    <td>{c.name || "Passkey"}</td>
                    <td>{formatWhen(c.createdAt)}</td>
                    <td>
                      <Button
                        size="sm"
                        variant="outline-danger"
                        disabled={busy}
                        onClick={() => handleDelete(c.id)}
                      >
                        Remove
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </Table>
          )}
        </>
      )}
    </section>
  );
}
