import { useEffect, useState } from "react";
import { Button, ButtonGroup, ButtonToolbar } from "react-bootstrap";
import { FaChevronLeft, FaChevronRight } from "react-icons/fa6";
import apiservice from "../services/api.service";

/**
 * Paginated page view: optional background (PDF raster / blank) + transparent
 * .rm SVG ink overlay from GET .../page/:n/overlay.svg (rmc / lines2svg).
 */
export default function AnnotatedPageViewer({
  docId,
  pageCount = 1,
  /** "pdf" shows payload background under the SVG; "notebook" uses white paper. */
  mode = "notebook",
  title = "",
}) {
  const total = Math.max(1, Number(pageCount) || 1);
  const [page, setPage] = useState(1);
  const [overlayUrl, setOverlayUrl] = useState(null);
  const [overlayError, setOverlayError] = useState(false);
  const [bgError, setBgError] = useState(false);

  useEffect(() => {
    setPage(1);
  }, [docId]);

  useEffect(() => {
    if (page > total) setPage(total);
  }, [page, total]);

  useEffect(() => {
    let cancelled = false;
    setOverlayUrl(null);
    setOverlayError(false);
    setBgError(false);
    const url = apiservice.getDocumentPageOverlayUrl(docId, page);
    fetch(url, { credentials: "same-origin" })
      .then((r) => {
        if (!r.ok) throw new Error(String(r.status));
        return r.blob();
      })
      .then((blob) => {
        if (cancelled) return;
        setOverlayUrl(URL.createObjectURL(blob));
      })
      .catch(() => {
        if (!cancelled) setOverlayError(true);
      });
    return () => {
      cancelled = true;
      setOverlayUrl((prev) => {
        if (prev) URL.revokeObjectURL(prev);
        return null;
      });
    };
  }, [docId, page]);

  const showBg = mode === "pdf";

  return (
    <div style={{ height: "95%", display: "flex", flexDirection: "column", minHeight: 0 }}>
      <ButtonToolbar className="gap-2 w-100 align-items-center mb-2 px-2">
        <ButtonGroup aria-label="Page">
          <Button
            size="sm"
            variant="outline-secondary"
            disabled={page <= 1}
            onClick={() => setPage((p) => Math.max(1, p - 1))}
          >
            <FaChevronLeft />
          </Button>
          <Button
            size="sm"
            variant="outline-secondary"
            disabled={page >= total}
            onClick={() => setPage((p) => Math.min(total, p + 1))}
          >
            <FaChevronRight />
          </Button>
        </ButtonGroup>
        <span className="text-muted small">
          Page {page} of {total}
        </span>
      </ButtonToolbar>

      <div style={{ flex: "1 1 auto", minHeight: 0, overflow: "auto", textAlign: "center", padding: "0 8px 16px" }}>
        <div
          style={{
            position: "relative",
            width: "100%",
            maxWidth: 900,
            margin: "0 auto",
            background: "#fff",
            border: "1px solid #ddd",
            borderRadius: 4,
            lineHeight: 0,
            overflow: "hidden",
          }}
        >
          {showBg && !bgError && (
            <img
              src={apiservice.getDocumentPageBackgroundUrl(docId, page)}
              alt=""
              aria-hidden
              onError={() => setBgError(true)}
              style={{ width: "100%", height: "auto", display: "block" }}
            />
          )}
          {(!showBg || bgError) && (
            <div
              aria-hidden
              style={{
                width: "100%",
                aspectRatio: "1404 / 1872",
                background: "#fff",
                display: showBg && !bgError ? "none" : "block",
              }}
            />
          )}
          {!overlayError && overlayUrl && (
            <img
              src={overlayUrl}
              alt={title ? `${title} page ${page} ink` : `Page ${page} ink`}
              style={{
                position: "absolute",
                inset: 0,
                width: "100%",
                height: "100%",
                objectFit: "fill",
                pointerEvents: "none",
              }}
            />
          )}
          {overlayError && (
            <p className="text-danger small" style={{ position: "absolute", inset: 0, padding: 12, lineHeight: 1.4 }}>
              Could not load ink overlay for page {page}.
            </p>
          )}
          {!overlayError && !overlayUrl && (
            <p className="text-muted small" style={{ position: "absolute", inset: 0, padding: 12, lineHeight: 1.4 }}>
              Loading ink…
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
