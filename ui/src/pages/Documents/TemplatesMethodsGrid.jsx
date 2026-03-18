import { useState, useEffect, useMemo } from "react";
import apiservice from "../../services/api.service";

const RATIO_W = 8.5;
const RATIO_H = 11;

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

function Card({ item, isMethod, fetchSvg }) {
  const [svgUrl, setSvgUrl] = useState(null);
  const [dimensions, setDimensions] = useState(null);

  useEffect(() => {
    let objectUrl = null;
    if (!item?.id || !fetchSvg) return;
    fetchSvg(item.id)
      .then((text) => {
        setDimensions(parseSvgDimensions(text));
        const blob = new Blob([text], { type: "image/svg+xml" });
        objectUrl = URL.createObjectURL(blob);
        setSvgUrl(objectUrl);
      })
      .catch(() => {});
    return () => {
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [item?.id, fetchSvg]);

  const orientation = dimensions
    ? dimensions.width > dimensions.height
      ? "Landscape"
      : "Portrait"
    : "";

  return (
    <div
      style={{
        border: isMethod ? "2px solid #0d6efd" : "1px solid #dee2e6",
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
          aspectRatio: `${RATIO_W} / ${RATIO_H}`,
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
        {svgUrl && (
          <img
            src={svgUrl}
            alt={item.name}
            style={{
              width: "100%",
              height: "100%",
              objectFit: "contain",
            }}
          />
        )}
      </div>
      <div style={{ marginTop: 6, textAlign: "center", fontSize: 12 }}>
        <div style={{ fontWeight: 600 }}>{item.name || item.id}</div>
        {orientation && <div>{orientation}</div>}
        {dimensions && (
          <div style={{ color: "#6c757d" }}>
            {dimensions.width} × {dimensions.height}
          </div>
        )}
        {isMethod && (
          <div style={{ color: "#0d6efd", fontSize: 11 }}>rm Method</div>
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
      ...templatesList.map((c) => ({ ...c, isMethod: false })),
      ...methodsList.map((c) => ({ ...c, isMethod: true })),
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
            isMethod={item.isMethod}
            fetchSvg={item.isMethod ? fetchMethod : fetchTemplate}
          />
        ))}
      </div>
    </div>
  );
}
