import { useEffect, useMemo, useState } from "react";
import Navbar from "react-bootstrap/Navbar";
import NameTag from "../../components/NameTag";
import constants from "../../common/constants";
import apiservice from "../../services/api.service";

/**
 * Renders an EPUB as a native website: backend serves unpacked HTML/CSS/assets by path;
 * the iframe loads the first spine item and relative links work like a normal site.
 */
export default function EpubViewer({ file, onSelect }) {
  const { data } = file;
  const epubBase = useMemo(() => {
    if (typeof window === "undefined") return "";
    return `${window.location.origin}${constants.ROOT_URL}/documents/${data?.id}/epub/`;
  }, [data?.id]);

  const [manifest, setManifest] = useState(null);
  const [currentPath, setCurrentPath] = useState(null);

  useEffect(() => {
    if (!data?.id) return;
    let cancelled = false;
    apiservice
      .getDocumentMetadata(data.id)
      .catch(() => {}); // warm auth; ignore
    fetch(`${constants.ROOT_URL}/documents/${data.id}/epub/manifest`, { credentials: "same-origin" })
      .then((r) => r.json())
      .then((m) => {
        if (cancelled) return;
        setManifest(m);
        if (m?.spine?.length) setCurrentPath(m.spine[0]);
      })
      .catch(() => {
        if (cancelled) return;
        setManifest({ spine: [] });
      });
    return () => {
      cancelled = true;
    };
  }, [data?.id]);

  const items = useMemo(() => {
    const spine = manifest?.spine ?? [];
    return spine.map((p) => {
      const parts = String(p).split("/");
      const label = parts[parts.length - 1] || p;
      return { path: p, label };
    });
  }, [manifest]);

  return (
    <>
      <Navbar style={{ marginLeft: "-12px" }}>
        {file && (
          <div>
            <NameTag node={file} onSelect={onSelect} />
          </div>
        )}
      </Navbar>

      <div
        style={{
          flex: 1,
          minHeight: 0,
          display: "flex",
          overflow: "hidden",
          background: "#fff",
        }}
      >
        <div style={{ width: 280, borderRight: "1px solid #dee2e6", overflow: "auto", background: "#f8f9fa" }}>
          <div style={{ padding: "10px 12px", fontWeight: 700, fontSize: 12, borderBottom: "1px solid #dee2e6" }}>
            Contents
          </div>
          {items.length === 0 && (
            <div style={{ padding: 12, fontSize: 12, color: "#6c757d" }}>
              Loading…
            </div>
          )}
          {items.map((it) => (
            <button
              key={it.path}
              type="button"
              onClick={() => setCurrentPath(it.path)}
              style={{
                width: "100%",
                textAlign: "left",
                padding: "10px 12px",
                border: "none",
                background: it.path === currentPath ? "#e9ecef" : "transparent",
                cursor: "pointer",
                fontSize: 12,
              }}
              title={it.path}
            >
              {it.label}
            </button>
          ))}
        </div>
        <div style={{ flex: 1, minWidth: 0, display: "flex", flexDirection: "column" }}>
          {data?.id && currentPath && (
            <iframe
              title={file?.data?.name || "EPUB"}
              src={`${epubBase}${currentPath}`}
              style={{
                flex: 1,
                minHeight: 0,
                width: "100%",
                border: "none",
              }}
            />
          )}
        </div>
      </div>
    </>
  );
}
