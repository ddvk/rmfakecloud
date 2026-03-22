import { useRef, useEffect, useState } from "react";
import { Button, Dropdown, ButtonToolbar } from "react-bootstrap";
import Navbar from "react-bootstrap/Navbar";
import { AiOutlineDownload } from "react-icons/ai";
import { BsBoxArrowUpRight } from "react-icons/bs";
import constants from "../../common/constants";

import apiservice from "../../services/api.service";
import NameTag from "../../components/NameTag";

export default function FileViewer({ file, onSelect }) {
  const { data } = file;

  const downloadUrl = `${constants.ROOT_URL}/documents/${file.id}`;

  const [height, setHeight] = useState(400);
  const [notebookMeta, setNotebookMeta] = useState(null);
  const [pdfBlobUrl, setPdfBlobUrl] = useState(null);
  const [pdfLoadError, setPdfLoadError] = useState(false);
  const pdfContainerRef = useRef(null);

  useEffect(() => {
    if (data?.type !== "pdf" || !pdfContainerRef.current) return;
    const el = pdfContainerRef.current;
    const ro = new ResizeObserver((entries) => {
      const e = entries[0];
      const h = e.contentBoxSize?.[0]?.blockSize ?? e.contentRect.height;
      setHeight(Math.max(400, h));
    });
    ro.observe(el);
    return () => ro.disconnect();
  }, [data?.type, file?.id]);

  useEffect(() => {
    if (data?.type !== "pdf" || !file?.id) {
      setPdfBlobUrl((prev) => {
        if (prev) URL.revokeObjectURL(prev);
        return null;
      });
      setPdfLoadError(false);
      return;
    }
    let cancelled = false;
    setPdfLoadError(false);
    fetch(downloadUrl, { credentials: "same-origin" })
      .then((r) => {
        if (!r.ok) throw new Error(String(r.status));
        return r.blob();
      })
      .then((blob) => {
        if (cancelled) return;
        const url = URL.createObjectURL(blob);
        setPdfBlobUrl((prev) => {
          if (prev) URL.revokeObjectURL(prev);
          return url;
        });
      })
      .catch(() => {
        if (!cancelled) setPdfLoadError(true);
      });
    return () => {
      cancelled = true;
      setPdfBlobUrl((prev) => {
        if (prev) URL.revokeObjectURL(prev);
        return null;
      });
    };
  }, [data?.type, file?.id, downloadUrl]);

  useEffect(() => {
    const needsMeta =
      file?.id &&
      (data?.type === "notebook" ||
        data?.type === "TODO" ||
        (data?.hasWritings && data?.type !== "pdf" && data?.type !== "epub"));
    if (needsMeta) {
      apiservice
        .getDocumentMetadata(file.id)
        .then(setNotebookMeta)
        .catch(() => setNotebookMeta({ pageCount: 0, hasWritings: !!data?.hasWritings }));
    } else {
      setNotebookMeta(null);
    }
  }, [file?.id, data?.type, data?.hasWritings]);

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
    // Open the PDF stream directly (same URL as inline preview), not /view-pdf/ wrapper.
    window.open(`${window.location.origin}${downloadUrl}`, "_blank", "noopener,noreferrer");
  };

  const isPdfView = data?.type === "pdf";
  const isNotebookView =
    data?.type === "notebook" ||
    data?.type === "TODO" ||
    (data?.hasWritings && data?.type !== "pdf" && data?.type !== "epub");

  const notebookPageCount =
    notebookMeta &&
    (notebookMeta.pageCount > 0
      ? notebookMeta.pageCount
      : notebookMeta.hasWritings
        ? 1
        : 0);

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
          {isPdfView && (
            <Button
              size="sm"
              variant="secondary"
              onClick={onOpenInNewTab}
              title="Open in new tab"
              className="me-1"
            >
              <BsBoxArrowUpRight />
            </Button>
          )}
          {!isPdfView && (
            <Dropdown align="end">
              <Dropdown.Toggle size="sm" variant="secondary">
                <AiOutlineDownload />
              </Dropdown.Toggle>
              <Dropdown.Menu>
                <Dropdown.Item onClick={onDownloadPdf}>Download PDF</Dropdown.Item>
                <Dropdown.Item onClick={onDownloadRmdoc}>Download .rmdoc</Dropdown.Item>
              </Dropdown.Menu>
            </Dropdown>
          )}
        </ButtonToolbar>
      </Navbar>

      {file && isPdfView && (
        <div ref={pdfContainerRef} style={{ height: "95%", minHeight: 400 }}>
          {pdfLoadError && (
            <p className="text-danger p-2">Could not load PDF preview.</p>
          )}
          {!pdfLoadError && pdfBlobUrl && (
            <iframe
              title={data.name || "PDF"}
              src={pdfBlobUrl}
              width="100%"
              height={height}
              style={{ border: "none", minHeight: 400 }}
            />
          )}
          {!pdfLoadError && !pdfBlobUrl && (
            <p className="text-muted p-2">Loading PDF…</p>
          )}
        </div>
      )}

      {file && isNotebookView && (
        <div
          style={{
            height: "95%",
            overflow: "auto",
            textAlign: "center",
            padding: "0 8px",
          }}
        >
          {notebookMeta && notebookPageCount > 0 ? (
            Array.from({ length: notebookPageCount }, (_, i) => {
              const pageNum = i + 1;
              return (
                <img
                  key={pageNum}
                  src={apiservice.getDocumentPagePngUrl(file.id, pageNum)}
                  alt={`${data.name} page ${pageNum}`}
                  loading="lazy"
                  decoding="async"
                  style={{
                    width: "100%",
                    maxWidth: 900,
                    height: "auto",
                    display: "block",
                    margin: "0 auto 16px",
                    border: "1px solid #ddd",
                    borderRadius: 4,
                  }}
                />
              );
            })
          ) : (
            <p style={{ padding: 16 }}>
              {notebookMeta
                ? "No pages to preview."
                : "Loading notebook preview…"}
            </p>
          )}
        </div>
      )}

      {file && !isPdfView && !isNotebookView && (
        <div style={{ height: "95%", padding: 16 }}>
          <p className="text-muted">Preview not available for this document type.</p>
          <a href={downloadUrl} target="_blank" rel="noopener noreferrer">Open document</a>
        </div>
      )}
    </>
  );
}
