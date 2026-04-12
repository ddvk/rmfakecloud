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
 * Grid thumbnail: optional background PNG (template / PDF page) with strokes PNG on top.
 * Strokes from the API are white-backed; multiply blends white away so lines sit on the background.
 */
export default function DocumentPageThumb({
  docId,
  pageNum = 1,
  alt = "",
  /** When false, only the full-page PNG is shown (no background layer). */
  layered = true,
  /** Shown if the foreground PNG fails to load. */
  fallback = null,
}) {
  const [bgState, setBgState] = useState(layered ? "pending" : "fail");
  const [fgFailed, setFgFailed] = useState(false);

  if (fgFailed) {
    return fallback ? <span style={{ display: "inline-flex" }}>{fallback}</span> : null;
  }

  const showBg = layered && bgState === "ok";
  const blendMode = showBg ? "multiply" : "normal";

  return (
    <div style={frameStyle} className="document-page-thumb" aria-label={alt || undefined}>
      {layered && (
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
      )}
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
