import { useEffect, useState } from "react";
import apiservice from "../services/api.service";

/**
 * Loads one notebook page as SVG (GET .../page/n/overlay.svg) with credentials and displays it.
 * Replaces PNG raster preview; strokes render as vector on a white backing.
 */
export default function NotebookPageSvg({ docId, pageNum, label }) {
  const [objectUrl, setObjectUrl] = useState(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    let cancelled = false;
    setObjectUrl(null);
    setError(false);
    const url = apiservice.getDocumentPageOverlayUrl(docId, pageNum);
    fetch(url, { credentials: "same-origin" })
      .then((r) => {
        if (!r.ok) throw new Error(String(r.status));
        return r.blob();
      })
      .then((blob) => {
        if (cancelled) return;
        const ou = URL.createObjectURL(blob);
        setObjectUrl(ou);
      })
      .catch(() => {
        if (!cancelled) setError(true);
      });
    return () => {
      cancelled = true;
      setObjectUrl((prev) => {
        if (prev) URL.revokeObjectURL(prev);
        return null;
      });
    };
  }, [docId, pageNum]);

  if (error) {
    return (
      <p className="text-danger small" style={{ padding: 8 }}>
        Could not load page {pageNum} (SVG).
      </p>
    );
  }
  if (!objectUrl) {
    return (
      <div
        style={{
          minHeight: 120,
          margin: "0 auto 16px",
          maxWidth: 900,
          background: "#f5f5f5",
          borderRadius: 4,
        }}
      >
        Loading…
      </div>
    );
  }

  return (
    <div
      style={{
        width: "100%",
        maxWidth: 900,
        margin: "0 auto 16px",
        background: "#fff",
        border: "1px solid #ddd",
        borderRadius: 4,
        lineHeight: 0,
      }}
    >
      <img
        src={objectUrl}
        alt={label || `Page ${pageNum}`}
        style={{
          width: "100%",
          height: "auto",
          display: "block",
          verticalAlign: "top",
        }}
      />
    </div>
  );
}
