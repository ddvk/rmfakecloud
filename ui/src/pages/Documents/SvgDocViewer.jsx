import { useState, useEffect, useRef } from "react";
import Navbar from "react-bootstrap/Navbar";
import NameTag from "../../components/NameTag";

/**
 * Shared viewer for SVG documents (templates and rm Methods).
 * Renders the same layout for both so methods look like normal templates.
 */
export default function SvgDocViewer({ file, onSelect, fetchSvg, errorMessage = "Failed to load" }) {
  const { data } = file;
  const [svgUrl, setSvgUrl] = useState(null);
  const [error, setError] = useState(null);
  const containerRef = useRef(null);

  useEffect(() => {
    let objectUrl = null;
    if (!fetchSvg) return;
    fetchSvg(data.id)
      .then((svgText) => {
        const blob = new Blob([svgText], { type: "image/svg+xml" });
        objectUrl = URL.createObjectURL(blob);
        setSvgUrl(objectUrl);
        setError(null);
      })
      .catch((err) => {
        setError(err?.message || errorMessage);
      });
    return () => {
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [data.id, fetchSvg, errorMessage]);

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
        ref={containerRef}
        style={{
          flex: 1,
          minHeight: 0,
          overflow: "auto",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          background: "#f5f5f5",
        }}
      >
        {error && <div className="text-danger p-3">{error}</div>}
        {svgUrl && (
          <img
            src={svgUrl}
            alt={data.name}
            style={{
              maxWidth: "100%",
              maxHeight: "100%",
              objectFit: "contain",
            }}
          />
        )}
      </div>
    </>
  );
}
