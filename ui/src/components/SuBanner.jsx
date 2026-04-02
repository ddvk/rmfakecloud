import { useState } from "react";
import { Alert, Button } from "react-bootstrap";
import { toast } from "react-toastify";
import apiService from "../services/api.service";
import { useAuthState } from "../common/useAuthContext";

export default function SuBanner() {
  const { state: { user }, dispatch } = useAuthState();
  const [busy, setBusy] = useState(false);

  if (!user || !user.SuBy) return null;

  const leaveSu = async () => {
    if (busy) return;
    setBusy(true);
    try {
      const restored = await apiService.leaveSu();
      dispatch({ type: "LOGIN_SUCCESS", payload: { user: restored } });
      window.location.replace("/documents");
    } catch (err) {
      toast.error(`Error: ${err.message || String(err)}`);
      setBusy(false);
    }
  };

  return (
    <Alert variant="warning" className="m-0 rounded-0 d-flex align-items-center justify-content-between">
      <div>
        <strong>SU active:</strong> acting as <strong>{user.UserID}</strong> (from {user.SuBy})
      </div>
      <Button size="sm" variant="dark" disabled={busy} onClick={leaveSu}>
        {busy ? "Leaving..." : "Leave su"}
      </Button>
    </Alert>
  );
}
