import { useState, useEffect, useRef, useMemo } from "react";
import { Button, ButtonGroup, Dropdown, ButtonToolbar } from "react-bootstrap";
import Navbar from 'react-bootstrap/Navbar';
import { FaChevronRight, FaChevronLeft } from "react-icons/fa6";
import { AiOutlineDownload } from "react-icons/ai";
import { BsBoxArrowUpRight } from "react-icons/bs";
import constants from "../../common/constants";

import apiservice from "../../services/api.service"
import NameTag from "../../components/NameTag"

import { pdfjs, Document, Page } from "react-pdf";

export default function FileViewer({ file, onSelect }) {
  const { data } = file;

  const downloadUrl = `${constants.ROOT_URL}/documents/${file.id}`;

  const [hasWritings, setHasWritings] = useState(false);
  const [page, setPage] = useState(1);
  const [pages, setPages] = useState(1);
  const [height, setHeight] = useState(100);
  const parent = useRef(null);
  const [overlaySvg, setOverlaySvg] = useState(null);

  const onLoadSuccess = (pdf) => {
    setPage(1);
    setPages(pdf.numPages);
  };
  const onPrev = () => setPage((p) => Math.max(p - 1, 1));
  const onNext = () => setPage((p) => Math.min(p + 1, pages));

  // Determine if the PDF has handwriting (.rm files).
  useEffect(() => {
    if (!file || data?.type !== "pdf") return;
    apiservice
      .getDocumentMetadata(file.id)
      .then((m) => {
        setHasWritings(Boolean(m?.hasWritings));
        if (typeof m?.pageCount === "number" && m.pageCount > 0) setPages(m.pageCount);
      })
      .catch(() => {});
  }, [file?.id, data?.type]);

  useEffect(() => {
    if (!parent.current) return;
    const ro = new ResizeObserver((e) => setHeight(e[0].contentBoxSize[0].blockSize));
    ro.observe(parent.current);
    return () => ro.disconnect();
  }, []);

  // For handwriting PDFs: fetch the vector overlay SVG for the current page.
  useEffect(() => {
    if (!file || data?.type !== "pdf") return;
    if (!hasWritings) return;
    apiservice
      .getDocumentPageOverlaySvg(file.id, page)
      .then((t) => setOverlaySvg(t))
      .catch(() => setOverlaySvg(null));
  }, [file?.id, data?.type, hasWritings, page]);

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

  const options = useMemo(() => ({ worker: new pdfjs.PDFWorker() }), [pdfjs]);

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
          {data?.type === "pdf" && hasWritings && pages > 1 && (
            <ButtonGroup>
              <Button size="sm" variant="outline-secondary" onClick={onPrev}>
                <FaChevronLeft />
              </Button>
              <Button size="sm" variant="outline-secondary" onClick={onNext}>
                <FaChevronRight />
              </Button>
              <span style={{ margin: "0 10px" }}>Page: {page} of {pages}</span>
            </ButtonGroup>
          )}
        </ButtonToolbar>
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
      </Navbar>

      {file && data?.type === "pdf" && !hasWritings && (
        // Use the browser's PDF plugin (Adobe if installed) when there is no handwriting.
        <div ref={parent} style={{ height: "95%" }}>
          <object
            data={downloadUrl}
            type="application/pdf"
            style={{ width: "100%", height: height, minHeight: 400 }}
          >
            <a href={downloadUrl}>Open PDF</a>
          </object>
        </div>
      )}

      {file && data?.type === "pdf" && hasWritings && (
        // Handwriting present: background PNG + vector overlay as separate layer.
        <div ref={parent} style={{ height: "95%", display: "flex", alignItems: "center", justifyContent: "center" }}>
          <div style={{ position: "relative", height: height, maxWidth: "100%", maxHeight: "100%" }}>
            <img
              src={apiservice.getDocumentPageBackgroundUrl(file.id, page)}
              alt={data.name}
              style={{ height: "100%", width: "auto", maxWidth: "100%", display: "block" }}
            />
            {overlaySvg && (
              <div
                style={{ position: "absolute", inset: 0 }}
                dangerouslySetInnerHTML={{ __html: overlaySvg }}
              />
            )}
          </div>
        </div>
      )}

      {file && data?.type !== "pdf" && (
        <div ref={parent} style={{ height: "95%" }}>
          <Document file={downloadUrl} onLoadSuccess={onLoadSuccess} options={options}>
            <Page
              pageNumber={page}
              height={height}
              renderAnnotationLayer={false}
              renderTextLayer={false}
            />
          </Document>
        </div>
      )}
    </>
  );
}
