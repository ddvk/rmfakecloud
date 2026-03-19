import { useRef, useEffect, useState } from "react";
import { Button, Dropdown, ButtonToolbar } from "react-bootstrap";
import Navbar from 'react-bootstrap/Navbar';
import { AiOutlineDownload } from "react-icons/ai";
import { BsBoxArrowUpRight } from "react-icons/bs";
import constants from "../../common/constants";

import apiservice from "../../services/api.service";
import NameTag from "../../components/NameTag";

export default function FileViewer({ file, onSelect }) {
  const { data } = file;

  const downloadUrl = `${constants.ROOT_URL}/documents/${file.id}`;

  const [height, setHeight] = useState(400);
  const parent = useRef(null);

  useEffect(() => {
    if (!parent.current) return;
    const ro = new ResizeObserver((e) => setHeight(e[0].contentBoxSize[0].blockSize));
    ro.observe(parent.current);
    return () => ro.disconnect();
  }, []);

  const triggerDownload = (blob, filename) => {
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    a.remove();
    window.URL.revokeObjectURL(url);
  };

  const onDownloadPdf = () => {
    apiservice.download(data.id)
      .then((blob) => triggerDownload(blob, data.name + ".pdf"))
      .catch(() => {});
  };

  const onDownloadRmdoc = () => {
    apiservice.download(data.id, "rmdoc")
      .then((blob) => triggerDownload(blob, data.name + ".rmdoc"))
      .catch(() => {});
  };

  const onOpenInNewTab = () => {
    const base = `${window.location.origin}/view-pdf/${file.id}`;
    const query = new URLSearchParams();
    if (data.name) query.set("name", data.name);
    const url = query.toString() ? `${base}?${query.toString()}` : base;
    window.open(url, "_blank", "noopener,noreferrer");
  };

  // PDFs and notebooks (exported as PDF) are served as raw application/pdf; render with browser/Adobe plugin only.
  const isPdfView = data?.type === "pdf" || data?.type === "notebook";

  return (
    <>
      <Navbar style={{ marginLeft: "-12px" }}>
        {file && (
          <div>
            <NameTag node={file} onSelect={onSelect} />
          </div>
        )}
      </Navbar>

      <Navbar>
        <ButtonToolbar className="gap-2">
          <div style={{ flex: 1 }} />
          <Button
            size="sm"
            variant="secondary"
            onClick={onOpenInNewTab}
            title="Open in new tab"
            className="me-1"
          >
            <BsBoxArrowUpRight />
          </Button>
          <Dropdown align="end">
            <Dropdown.Toggle size="sm" variant="secondary">
              <AiOutlineDownload />
            </Dropdown.Toggle>
            <Dropdown.Menu>
              <Dropdown.Item onClick={onDownloadPdf}>Download PDF</Dropdown.Item>
              <Dropdown.Item onClick={onDownloadRmdoc}>Download .rmdoc</Dropdown.Item>
            </Dropdown.Menu>
          </Dropdown>
        </ButtonToolbar>
      </Navbar>

      {file && isPdfView && (
        <div ref={parent} style={{ height: "95%" }}>
          <object
            data={downloadUrl}
            type="application/pdf"
            width="100%"
            height={height}
            style={{ minHeight: 400 }}
          >
            <p>Alternative text: <a href={downloadUrl}>Download PDF</a></p>
          </object>
        </div>
      )}

      {file && !isPdfView && (
        <div ref={parent} style={{ height: "95%", padding: 16 }}>
          <p className="text-muted">Preview not available for this document type.</p>
          <a href={downloadUrl} target="_blank" rel="noopener noreferrer">Open document</a>
        </div>
      )}
    </>
  );
}
