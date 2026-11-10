import { useCallback } from "react";
import apiservice from "../../services/api.service";
import SvgDocViewer from "./SvgDocViewer";

export default function TemplateViewer({ file, onSelect }) {
  const fetchSvg = useCallback((id) => apiservice.getTemplate(id), []);
  return (
    <SvgDocViewer
      file={file}
      onSelect={onSelect}
      fetchSvg={fetchSvg}
      errorMessage="Failed to load template"
    />
  );
}
