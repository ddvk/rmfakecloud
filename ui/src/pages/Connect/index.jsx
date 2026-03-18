import React, { useState, useLayoutEffect, useEffect, useRef } from "react";
import apiservice from "../../services/api.service";
import Stack from "react-bootstrap/Stack";
import Button from "react-bootstrap/Button";
import { FaRepeat } from "react-icons/fa6";

const CODE_VALIDITY_SEC = 5 * 60; // 300 seconds, match backend

function ExpiryPie({ expiresAt }) {
  const [now, setNow] = useState(() => Math.floor(Date.now() / 1000));
  useEffect(() => {
    const t = setInterval(() => setNow(Math.floor(Date.now() / 1000)), 1000);
    return () => clearInterval(t);
  }, []);
  if (!expiresAt || expiresAt <= now) {
    return (
      <div
        style={{
          width: 80,
          height: 80,
          borderRadius: "50%",
          background: "conic-gradient(#ccc 0deg 360deg)",
        }}
      />
    );
  }
  const remaining = Math.max(0, expiresAt - now);
  const ratio = 1 - remaining / CODE_VALIDITY_SEC; // 0 = full, 1 = empty
  const elapsedDeg = 360 * Math.min(1, ratio);
  return (
    <div
      style={{
        width: 80,
        height: 80,
        borderRadius: "50%",
        background: `conic-gradient(#e9ecef 0deg ${elapsedDeg}deg, #28a745 ${elapsedDeg}deg 360deg)`,
      }}
      title={`Expires in ${remaining}s`}
    />
  );
}

export default function CodeGenerator() {
  const [code, setCode] = useState("");
  const [expiresAt, setExpiresAt] = useState(null);
  const [error, setError] = useState("");
  const statusCheckRef = useRef(null);

  const fetchCode = async () => {
    setCode("");
    setExpiresAt(null);
    try {
      const data = await apiservice.getCode();
      setCode(data.code || "");
      setExpiresAt(data.expiresAt ?? null);
    } catch (e) {
      setError(e);
    }
  };

  useLayoutEffect(() => {
    fetchCode();
  }, []);

  // Every second: check if code was used or expired; refresh page if so
  useEffect(() => {
    if (!expiresAt) return;
    statusCheckRef.current = setInterval(async () => {
      const now = Math.floor(Date.now() / 1000);
      if (now >= expiresAt) {
        window.location.reload();
        return;
      }
      try {
        const st = await apiservice.getCodeStatus();
        if (!st.valid) {
          window.location.reload();
        }
      } catch (_) {
        window.location.reload();
      }
    }, 1000);
    return () => {
      if (statusCheckRef.current) clearInterval(statusCheckRef.current);
    };
  }, [expiresAt]);

  if (error) {
    return <div>{error.message}</div>;
  }

  return (
    <>
      <Stack gap={4} style={{ alignItems: "center", marginTop: "15vh" }}>
        <div className="p-2">
          <Button onClick={fetchCode}>
            <FaRepeat />
          </Button>
        </div>
        <div className="p-2">
          <h1 style={{ letterSpacing: "10px", fontSize: "36pt" }}>{code}</h1>
        </div>
        <div className="p-2">
          <ExpiryPie expiresAt={expiresAt} />
        </div>
      </Stack>
    </>
  );
}
