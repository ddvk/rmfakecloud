import { useState } from "react";
import apiservice from "../services/api.service";

const frameStyle = {
  width: 68,
  height: 88,
  borderRadius: 3,
  border: "1px solid #dee2e6",
  background: "#fff",
  overflow: "hidden",
  position: "relative",
  flexShrink: 0,
  isolation: "isolate",
};

const layerImg = {
  position: "absolute",
  inset: 0,
  width: "100%",
  height: "100%",
  objectFit: "cover",
  pointerEvents: "none",
};

/**
 * Grid thumbnail. Prefers the server composite thumb (device-like 384×512).
 * Optional layered mode keeps bg+ink multiply for debugging / fallback.
 */
export default function DocumentPageThumb({
  docId,
  pageNum = 1,
  alt = "",
  /** When true, use client-side bg+ink layers instead of server thumb. */
  layered = false,
  /** Shown if the thumbnail PNG fails to load. */
  fallback = null,
}) {
  const [bgState, setBgState] = useState(layered ? "pending" : "fail");
  const [fgFailed, setFgFailed] = useState(false);

  if (fgFailed) {
    return fallback ? <span style={{ display: "inline-flex" }}>{fallback}</span> : null;
  }

  if (!layered) {
    return (
      <div style={frameStyle} className="document-page-thumb" aria-label={alt || undefined}>
        <img
          src={apiservice.getDocumentPageThumbUrl(docId, pageNum)}
          alt={alt}
          loading="lazy"
          decoding="async"
          onError={() => setFgFailed(true)}
          style={{ ...layerImg, position: "relative", zIndex: 1 }}
        />
      </div>
    );
  }

  const showBg = bgState === "ok";
  const blendMode = showBg ? "multiply" : "normal";

  return (
    <div style={frameStyle} className="document-page-thumb" aria-label={alt || undefined}>
      <img
        src={apiservice.getDocumentPageBackgroundUrl(docId, pageNum)}
        alt=""
        aria-hidden
        loading="lazy"
        decoding="async"
        onLoad={() => setBgState("ok")}
        onError={() => setBgState("fail")}
        style={{
          ...layerImg,
          display: showBg ? "block" : "none",
          zIndex: 0,
        }}
      />
      <img
        src={apiservice.getDocumentPagePngUrl(docId, pageNum)}
        alt={alt}
        loading="lazy"
        decoding="async"
        onError={() => setFgFailed(true)}
        style={{
          ...layerImg,
          zIndex: 1,
          mixBlendMode: blendMode,
        }}
      />
    </div>
  );
}
