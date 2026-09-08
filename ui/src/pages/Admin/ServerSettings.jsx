import { useEffect, useState } from "react";
import { Alert, Button, Form, Stack } from "react-bootstrap";
import { toast } from "react-toastify";
import apiService from "../../services/api.service";

/**
 * Admin-only rendering tooling paths. Persisted on the server under DATADIR
 * (overrides env). RMFAKECLOUD_ALLOW_SU is intentionally not shown here.
 */
export default function ServerSettings() {
  const [rmcSrc, setRmcSrc] = useState("");
  const [envName, setEnvName] = useState("RMFAKECLOUD_RMC_SRC");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [loadError, setLoadError] = useState("");

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    apiService
      .getServerSettings()
      .then((s) => {
        if (cancelled) return;
        setRmcSrc(s.rmcSrc || "");
        if (s.rmcSrcEnv) setEnvName(s.rmcSrcEnv);
        setLoadError("");
      })
      .catch((e) => {
        if (!cancelled) setLoadError(e.message || String(e));
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const save = async (e) => {
    e.preventDefault();
    if (saving) return;
    setSaving(true);
    try {
      const s = await apiService.updateServerSettings({ rmcSrc: rmcSrc.trim() });
      setRmcSrc(s.rmcSrc || "");
      toast.success("Server settings saved");
    } catch (err) {
      toast.error(`Error: ${err.message || String(err)}`);
    } finally {
      setSaving(false);
    }
  };

  return (
    <section className="mb-4">
      <h3>Rendering</h3>
      <p className="text-muted small">
        Paths used for high-fidelity notebook / v6 <code>.rm</code> conversion.
        Saved under the server data directory (overrides the environment variable).
      </p>
      {loadError && <Alert variant="danger">{loadError}</Alert>}
      {loading ? (
        <p>Loading…</p>
      ) : (
        <Form onSubmit={save}>
          <Stack gap={3}>
            <Form.Group controlId="rmcSrc">
              <Form.Label>
                <code>{envName}</code> — rmc source directory
              </Form.Label>
              <Form.Control
                type="text"
                value={rmcSrc}
                onChange={(ev) => setRmcSrc(ev.target.value)}
                placeholder="/path/to/rmc-main/src"
                autoComplete="off"
                spellCheck={false}
              />
              <Form.Text muted>
                Directory that contains the <code>rmc</code> Python package (usually{" "}
                <code>…/rmc-main/src</code>). Used as <code>python3 -m rmc.cli</code> for
                v6 SVG/PNG when the <code>rmc</code> binary is not on PATH. Leave empty to
                clear the override.
              </Form.Text>
            </Form.Group>
            <div>
              <Button type="submit" disabled={saving}>
                {saving ? "Saving…" : "Save"}
              </Button>
            </div>
          </Stack>
        </Form>
      )}
    </section>
  );
}
