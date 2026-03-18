import { useState, useEffect, useMemo } from "react";
import apiservice from "../../services/api.service";

// Aspect ratio presets (inches): [label, width, height]
const ASPECT_PRESETS = [
  { id: "rmpp portrait", label: "rmpp portrait", w: 8.5, h: 11 },
  { id: "rmpp landscape", label: "rmpp landscape", w: 11, h: 8.5 },
  { id: "rmppm portrait", label: "rmppm portrait", w: 5, h: 8 },
  { id: "rmppm landscape", label: "rmppm landscape", w: 8, h: 5 },
];

function parseSvgDimensions(svgText) {
  if (!svgText || typeof svgText !== "string") return null;
  const viewBoxMatch = svgText.match(/viewBox\s*=\s*["']?\s*[\d.]+\s+[\d.]+\s+([\d.]+)\s+([\d.]+)/);
  if (viewBoxMatch) {
    const w = Math.round(parseFloat(viewBoxMatch[1]));
    const h = Math.round(parseFloat(viewBoxMatch[2]));
    return { width: w, height: h };
  }
  const wMatch = svgText.match(/width\s*=\s*["']?(\d+)/);
  const hMatch = svgText.match(/height\s*=\s*["']?(\d+)/);
  if (wMatch && hMatch) {
    return { width: parseInt(wMatch[1], 10), height: parseInt(hMatch[1], 10) };
  }
  return null;
}

function getPresetForDimensions(dimensions) {
  if (!dimensions || !dimensions.width || !dimensions.height) return ASPECT_PRESETS[0];
  const w = dimensions.width;
  const h = dimensions.height;
  const isPortrait = w < h;
  const aspect = w / h;
  if (isPortrait) {
    if (aspect >= 0.7 && aspect <= 0.85) return ASPECT_PRESETS[0]; // rmpp portrait 8.5/11
    if (aspect >= 0.5 && aspect < 0.7) return ASPECT_PRESETS[2]; // rmppm portrait 5/8
    return ASPECT_PRESETS[0];
  }
  if (aspect >= 1.2 && aspect <= 1.4) return ASPECT_PRESETS[1]; // rmpp landscape 11/8.5
  if (aspect >= 1.4 && aspect <= 1.8) return ASPECT_PRESETS[3]; // rmppm landscape 8/5
  return ASPECT_PRESETS[1];
}

function toPortraitPreset(preset) {
  if (!preset) return ASPECT_PRESETS[0];
  if (preset.id === "rmpp landscape") return ASPECT_PRESETS[0];
  if (preset.id === "rmppm landscape") return ASPECT_PRESETS[2];
  return preset;
}

function isSvgString(text) {
  if (typeof text !== "string") return false;
  const t = text.trim();
  return t.startsWith("<svg") || t.startsWith("<?xml");
}

function svgToScaledInline(svgText) {
  if (!svgText || !isSvgString(svgText)) return svgText;
  return svgText
    .replace(/<svg/, '<svg style="width:100%;height:100%;display:block;vertical-align:middle" preserveAspectRatio="xMidYMid meet"')
    .replace(/<\?xml[^>]*\?>/, "");
}

function Card({ item, isTemplate, fetchSvg }) {
  const [svgContent, setSvgContent] = useState(null);
  const [dimensions, setDimensions] = useState(null);

  useEffect(() => {
    if (!item?.id || !fetchSvg) return;
    fetchSvg(item.id)
      .then((text) => {
        setDimensions(parseSvgDimensions(text));
        setSvgContent(typeof text === "string" ? text : null);
      })
      .catch(() => {});
  }, [item?.id, fetchSvg]);

  const rawPreset = getPresetForDimensions(dimensions);
  const hasOrientationInMetadata = Boolean(item?.orientation && String(item.orientation).trim());
  const usePortraitAndNoLabel = isTemplate && !hasOrientationInMetadata;
  const preset = usePortraitAndNoLabel ? toPortraitPreset(rawPreset) : rawPreset;
  const orientation = dimensions
    ? dimensions.width > dimensions.height
      ? "Landscape"
      : "Portrait"
    : "";
  const isSvg = svgContent && isSvgString(svgContent);

  return (
    <div
      style={{
        border: isTemplate ? "2px solid #0d6efd" : "1px solid #dee2e6",
        borderRadius: 8,
        padding: 8,
        background: "#fff",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        minWidth: 0,
      }}
    >
      <div
        style={{
          aspectRatio: `${preset.w} / ${preset.h}`,
          width: "100%",
          maxWidth: 120,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          background: "#f5f5f5",
          borderRadius: 4,
          overflow: "hidden",
        }}
      >
        {isSvg ? (
          <div
            style={{
              width: "100%",
              height: "100%",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
            }}
            dangerouslySetInnerHTML={{ __html: svgToScaledInline(svgContent) }}
          />
        ) : svgContent ? (
          <img
            src={`data:image/svg+xml;utf8,${encodeURIComponent(svgContent)}`}
            alt={item.name}
            style={{
              width: "100%",
              height: "100%",
              objectFit: "contain",
            }}
          />
        ) : null}
      </div>
      <div style={{ marginTop: 6, textAlign: "center", fontSize: 12, color: "#000" }}>
        <div style={{ fontWeight: 600, color: "#000" }}>{item.name || item.id}</div>
        {(!usePortraitAndNoLabel && orientation) ? <div style={{ color: "#000" }}>{orientation}</div> : null}
        {dimensions && (
          <div style={{ color: "#000" }}>
            {dimensions.width} × {dimensions.height}
          </div>
        )}
        <div style={{ color: "#000", fontSize: 11 }}>{preset.label} ({preset.w}×{preset.h} in)</div>
        {isTemplate && (
          <div style={{ color: "#0d6efd", fontSize: 11 }}>Template</div>
        )}
      </div>
    </div>
  );
}

export default function TemplatesMethodsGrid({ templates, methods }) {
  const templatesList = useMemo(() => templates ?? [], [templates]);
  const methodsList = useMemo(() => methods ?? [], [methods]);
  const combined = useMemo(
    () => [
      ...templatesList.map((c) => ({ ...c, isTemplate: true })),
      ...methodsList.map((c) => ({ ...c, isTemplate: false })),
    ],
    [templatesList, methodsList]
  );

  const fetchTemplate = (id) => apiservice.getTemplate(id);
  const fetchMethod = (id) => apiservice.getMethod(id);

  return (
    <div
      style={{
        flex: 1,
        minHeight: 0,
        overflow: "auto",
        padding: 16,
      }}
    >
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fill, minmax(140px, 1fr))",
          gap: 16,
        }}
      >
        {combined.map((item) => (
          <Card
            key={item.id}
            item={item}
            isTemplate={item.isTemplate}
            fetchSvg={item.isTemplate ? fetchTemplate : fetchMethod}
          />
        ))}
      </div>
    </div>
  );
}
