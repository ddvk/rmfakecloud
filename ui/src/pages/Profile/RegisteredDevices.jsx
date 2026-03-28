import { useCallback, useEffect, useState } from "react";
import Table from "react-bootstrap/Table";
import Spinner from "react-bootstrap/Spinner";
import Button from "react-bootstrap/Button";
import Modal from "react-bootstrap/Modal";
import Form from "react-bootstrap/Form";
import { toast } from "react-toastify";

import apiservice from "../../services/api.service";

function formatWhen(iso) {
  if (!iso) return "—";
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
}

export default function RegisteredDevices() {
  const [devices, setDevices] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [tokenModal, setTokenModal] = useState({ show: false, token: "", deviceLabel: "" });
  const [busyId, setBusyId] = useState(null);

  const loadDevices = useCallback((opts = {}) => {
    const silent = Boolean(opts.silent);
    if (!silent) setLoading(true);
    return apiservice
      .listRegisteredDevices()
      .then((data) => {
        setDevices(Array.isArray(data.devices) ? data.devices : []);
        setError(null);
      })
      .catch((e) => {
        setError(e.message || String(e));
        setDevices([]);
      })
      .finally(() => {
        if (!silent) setLoading(false);
      });
  }, []);

  useEffect(() => {
    loadDevices();
  }, [loadDevices]);

  async function handleReissue(device) {
    setBusyId(device.deviceId);
    try {
      const { token } = await apiservice.reissueDeviceToken(device.deviceId);
      if (token) {
        setTokenModal({
          show: true,
          token,
          deviceLabel: device.deviceDesc || device.deviceId || "device",
        });
        toast.success("New device token issued — paste it on the tablet if needed.");
        loadDevices({ silent: true });
      }
    } catch (e) {
      toast.error(e.message || String(e));
    } finally {
      setBusyId(null);
    }
  }

  async function handleSyncAll() {
    setBusyId("__sync__");
    try {
      await apiservice.triggerSync();
      toast.success("Sync notification sent to connected devices.");
    } catch (e) {
      toast.error(e.message || String(e));
    } finally {
      setBusyId(null);
    }
  }

  function copyToken() {
    if (!tokenModal.token) return;
    navigator.clipboard.writeText(tokenModal.token).then(
      () => toast.success("Copied to clipboard"),
      () => toast.error("Could not copy")
    );
  }

  return (
    <>
      <h3 className="mt-4">Registered devices</h3>
      <p className="text-muted small">
        Devices stored when they pair with a code. You can request a new device token without a pairing code
        (web login required), or ping sync for tablets that are online.
      </p>
      {!loading && devices.length > 0 && (
        <div className="mb-2">
          <Button
            variant="outline-secondary"
            size="sm"
            disabled={busyId === "__sync__"}
            onClick={handleSyncAll}
          >
            {busyId === "__sync__" ? "Sending…" : "Notify sync (all online devices)"}
          </Button>
        </div>
      )}
      {loading && devices.length === 0 && (
        <Spinner animation="border" size="sm" role="status" className="me-2" />
      )}
      {error && <div className="alert alert-danger">{error}</div>}
      {!loading && !error && devices.length === 0 && (
        <p className="text-muted">No registered devices yet. Pair once with a code to appear here.</p>
      )}
      {!loading && devices.length > 0 && (
        <Table responsive striped bordered hover size="sm" className="mt-2">
          <thead>
            <tr>
              <th>Description</th>
              <th>Device ID</th>
              <th>Last seen</th>
              <th>Registered</th>
              <th style={{ width: "1%" }}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {devices.map((d) => (
              <tr key={d.deviceId || d.deviceDesc}>
                <td>{d.deviceDesc || "—"}</td>
                <td>
                  <code className="small">{d.deviceId || "—"}</code>
                </td>
                <td>{formatWhen(d.lastSeen)}</td>
                <td>{formatWhen(d.registeredAt)}</td>
                <td className="text-nowrap">
                  <Button
                    variant="outline-primary"
                    size="sm"
                    disabled={busyId !== null}
                    onClick={() => handleReissue(d)}
                  >
                    {busyId === d.deviceId ? "…" : "Re-issue token"}
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </Table>
      )}

      <Modal show={tokenModal.show} onHide={() => setTokenModal({ show: false, token: "", deviceLabel: "" })} size="lg">
        <Modal.Header closeButton>
          <Modal.Title>Device token — {tokenModal.deviceLabel}</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          <p className="small text-muted">
            Same kind of token as after pairing. Use on the tablet only if your client lets you paste or replace the stored token.
          </p>
          <Form.Control as="textarea" rows={6} readOnly value={tokenModal.token} className="font-monospace small" />
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={copyToken}>
            Copy
          </Button>
          <Button variant="primary" onClick={() => setTokenModal({ show: false, token: "", deviceLabel: "" })}>
            Close
          </Button>
        </Modal.Footer>
      </Modal>
    </>
  );
}
