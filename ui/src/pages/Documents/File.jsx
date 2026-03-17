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

// View mode: "annotated" = PDF with reMarkable drawings (react-pdf), "original" = PDF only via Adobe Embed
const VIEW_ANNOTATED = "annotated";
const VIEW_ORIGINAL = "original";

function useAdobeReady() {
  const [ready, setReady] = useState(() => typeof window !== "undefined" && window.AdobeDC);
  useEffect(() => {
    if (window.AdobeDC) {
      setReady(true);
      return;
    }
    const onReady = () => setReady(true);
    document.addEventListener("adobe_dc_view_sdk.ready", onReady);
    return () => document.removeEventListener("adobe_dc_view_sdk.ready", onReady);
  }, []);
  return ready;
}

export default function FileViewer({ file, onSelect }) {
  const { data } = file;

  const downloadUrl = `${constants.ROOT_URL}/documents/${file.id}`;
  const payloadPdfUrl = typeof window !== "undefined"
    ? `${window.location.origin}${constants.ROOT_URL}/documents/${file.id}?type=pdf&annotations=0`
    : "";

  const [viewMode, setViewMode] = useState(VIEW_ANNOTATED);
  const [page, setPage] = useState(1);
  const [pages, setPages] = useState(1);
  const [height, setHeight] = useState(100);
  const parent = useRef(null);
  const adobeContainerRef = useRef(null);
  const adobeViewRef = useRef(null);
  const adobeReady = useAdobeReady();
  const canShowOriginal = Boolean(constants.ADOBE_CLIENT_ID && adobeReady);

  const onLoadSuccess = (pdf) => {
    setPage(1);
    setPages(pdf.numPages);
  };
  const onPrev = () => setPage((p) => Math.max(p - 1, 1));
  const onNext = () => setPage((p) => Math.min(p + 1, pages));

  useEffect(() => {
    if (!parent.current) return;
    const ro = new ResizeObserver((e) => setHeight(e[0].contentBoxSize[0].blockSize));
    ro.observe(parent.current);
    return () => ro.disconnect();
  }, []);

  // Adobe Embed: show original PDF (no drawings) when viewMode is VIEW_ORIGINAL
  useEffect(() => {
    if (viewMode !== VIEW_ORIGINAL || !canShowOriginal || !adobeContainerRef.current || !payloadPdfUrl) return;
    const divId = "adobe-dc-view-" + file.id;
    adobeContainerRef.current.id = divId;
    const adobeDCView = new window.AdobeDC.View({
      clientId: constants.ADOBE_CLIENT_ID,
      divId,
      downloadWithCredentials: true,
    });
    adobeViewRef.current = adobeDCView;
    adobeDCView.previewFile({
      content: { location: { url: payloadPdfUrl } },
      metaData: { fileName: data.name || "document.pdf" },
    });
    return () => {
      adobeViewRef.current = null;
    };
  }, [viewMode, canShowOriginal, file.id, payloadPdfUrl, data.name]);

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
    if (viewMode === VIEW_ORIGINAL) query.set("original", "1");
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
          {viewMode === VIEW_ANNOTATED && pages > 1 && (
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
          {canShowOriginal && (
            <ButtonGroup size="sm">
              <Button
                variant={viewMode === VIEW_ANNOTATED ? "primary" : "outline-secondary"}
                onClick={() => setViewMode(VIEW_ANNOTATED)}
              >
                With drawings
              </Button>
              <Button
                variant={viewMode === VIEW_ORIGINAL ? "primary" : "outline-secondary"}
                onClick={() => setViewMode(VIEW_ORIGINAL)}
              >
                Original PDF
              </Button>
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

      {file && viewMode === VIEW_ANNOTATED && (
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

      {file && viewMode === VIEW_ORIGINAL && canShowOriginal && (
        <div ref={adobeContainerRef} style={{ height: "95%", minHeight: 400 }} />
      )}

      {file && viewMode === VIEW_ORIGINAL && !canShowOriginal && (
        <div className="p-3 text-warning">
          Set VITE_ADOBE_PDF_CLIENT_ID to use the Adobe viewer for the original PDF (no drawings).
        </div>
      )}
    </>
  );
}
