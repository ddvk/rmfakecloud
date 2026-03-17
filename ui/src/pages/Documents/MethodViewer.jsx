import { useCallback } from "react";
import apiservice from "../../services/api.service";
import SvgDocViewer from "./SvgDocViewer";

export default function MethodViewer({ file, onSelect }) {
  const fetchSvg = useCallback((id) => apiservice.getMethod(id), []);
  return (
    <SvgDocViewer
      file={file}
      onSelect={onSelect}
      fetchSvg={fetchSvg}
      errorMessage="Failed to load method"
    />
  );
}
