import { useEffect, useState } from "react";
import Table from "react-bootstrap/Table";
import Spinner from "react-bootstrap/Spinner";

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

  useEffect(() => {
    let cancelled = false;
    apiservice
      .listRegisteredDevices()
      .then((data) => {
        if (!cancelled) {
          setDevices(Array.isArray(data.devices) ? data.devices : []);
          setError(null);
        }
      })
      .catch((e) => {
        if (!cancelled) {
          setError(e.message || String(e));
          setDevices([]);
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <>
      <h3 className="mt-4">Registered devices</h3>
      <p className="text-muted small">
        Tablets that completed pairing with your account. New devices appear after you connect them with a pairing code.
      </p>
      {loading && (
        <Spinner animation="border" size="sm" role="status" className="me-2" />
      )}
      {error && <div className="alert alert-danger">{error}</div>}
      {!loading && !error && devices.length === 0 && (
        <p className="text-muted">No registered devices yet.</p>
      )}
      {!loading && devices.length > 0 && (
        <Table responsive striped bordered hover size="sm" className="mt-2">
          <thead>
            <tr>
              <th>Description</th>
              <th>Device ID</th>
              <th>Last seen</th>
              <th>Registered</th>
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
              </tr>
            ))}
          </tbody>
        </Table>
      )}
    </>
  );
}
