import { useState, useEffect, useRef } from "react";
import { useParams, useLocation } from "react-router-dom";
import constants from "../common/constants";

/**
 * Full-page Adobe PDF Embed viewer for opening a document in a new tab.
 * Route: /view-pdf/:docId?original=1&name=...
 */
export default function ViewPdf() {
  const { docId } = useParams();
  const { search } = useLocation();
  const containerRef = useRef(null);
  const [ready, setReady] = useState(() => typeof window !== "undefined" && window.AdobeDC);
  const [error, setError] = useState(null);

  const params = new URLSearchParams(search);
  const original = params.get("original") === "1" || params.get("original") === "true";
  const fileName = params.get("name") || "document.pdf";

  const pdfUrl =
    typeof window !== "undefined"
      ? `${window.location.origin}${constants.ROOT_URL}/documents/${docId}${original ? "?type=pdf&annotations=0" : ""}`
      : "";

  useEffect(() => {
    if (window.AdobeDC) {
      setReady(true);
      return;
    }
    const onReady = () => setReady(true);
    document.addEventListener("adobe_dc_view_sdk.ready", onReady);
    return () => document.removeEventListener("adobe_dc_view_sdk.ready", onReady);
  }, []);

  const initializedRef = useRef(false);
  useEffect(() => {
    if (!docId || !pdfUrl || !constants.ADOBE_CLIENT_ID || !ready || !containerRef.current || initializedRef.current) return;
    initializedRef.current = true;
    const divId = "adobe-dc-view-full";
    containerRef.current.id = divId;
    try {
      const adobeDCView = new window.AdobeDC.View({
        clientId: constants.ADOBE_CLIENT_ID,
        divId,
        downloadWithCredentials: true,
      });
      adobeDCView.previewFile({
        content: { location: { url: pdfUrl } },
        metaData: { fileName },
      });
    } catch (err) {
      setError(err?.message || "Failed to load PDF viewer");
    }
  }, [docId, pdfUrl, ready, fileName]);

  if (!constants.ADOBE_CLIENT_ID) {
    return (
      <div style={{ padding: 24, textAlign: "center" }}>
        <p className="text-warning">
          Set VITE_ADOBE_PDF_CLIENT_ID to view PDFs in a new tab with the Adobe viewer.
        </p>
        <p>
          <a href={`${typeof window !== "undefined" ? window.location.origin : ""}${constants.ROOT_URL}/documents/${docId}`}>
            Open PDF in browser
          </a>
        </p>
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ padding: 24, textAlign: "center" }}>
        <p className="text-danger">{error}</p>
      </div>
    );
  }

  return (
    <div
      ref={containerRef}
      style={{
        position: "fixed",
        inset: 0,
        minHeight: 0,
        background: "#f0f0f0",
      }}
    />
  );
}
