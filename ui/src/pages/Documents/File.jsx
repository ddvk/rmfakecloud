import { useRef, useEffect, useState } from "react";
import { Button, ButtonGroup, Dropdown, ButtonToolbar } from "react-bootstrap";
import Navbar from "react-bootstrap/Navbar";
import { AiOutlineDownload } from "react-icons/ai";
import { BsBoxArrowUpRight } from "react-icons/bs";
import { FaChevronLeft, FaChevronRight } from "react-icons/fa6";
import constants from "../../common/constants";

import apiservice from "../../services/api.service";
import NameTag from "../../components/NameTag";
import NotebookPageSvg from "../../components/NotebookPageSvg";
import { RmLinesRenderer } from "../../components/RmLinesRenderer";
import { tryDocumentEnvironment } from "../../common/blobapi";

export default function FileViewer({ file, onSelect }) {
  const { data } = file;

  const downloadUrl = `${constants.ROOT_URL}/documents/${file.id}`;

  const [height, setHeight] = useState(400);
  const [notebookMeta, setNotebookMeta] = useState(null);
  const [pdfBlobUrl, setPdfBlobUrl] = useState(null);
  const [pdfLoadError, setPdfLoadError] = useState(false);
  const pdfContainerRef = useRef(null);

  /** Sync 1.5+ blob tree: client-side librm_lines when available */
  const [rmLinesEnv, setRmLinesEnv] = useState(null);
  const [rmLinesChecked, setRmLinesChecked] = useState(false);
  const [rmPage, setRmPage] = useState(1);

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

  const isPdfView = data?.type === "pdf";
  const isNotebookView =
    data?.type === "notebook" ||
    data?.type === "TODO" ||
    (data?.hasWritings && data?.type !== "pdf" && data?.type !== "epub");

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

  useEffect(() => {
    setRmPage(1);
  }, [file?.id]);

  useEffect(() => {
    if (!isNotebookView || !file?.id) {
      setRmLinesEnv(null);
      setRmLinesChecked(false);
      return;
    }
    let cancelled = false;
    setRmLinesChecked(false);
    setRmLinesEnv(null);
    (async () => {
      const env = await tryDocumentEnvironment(file.id);
      if (cancelled) return;
      setRmLinesEnv(env);
      setRmLinesChecked(true);
    })();
    return () => {
      cancelled = true;
    };
  }, [isNotebookView, file?.id]);

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
    window.open(`${window.location.origin}${downloadUrl}`, "_blank", "noopener,noreferrer");
  };

  const notebookPageCount =
    notebookMeta &&
    (notebookMeta.pageCount > 0
      ? notebookMeta.pageCount
      : notebookMeta.hasWritings
        ? 1
        : 0);

  /** Prefer server metadata; before it loads, blob tree .rm count is a usable fallback for librm_lines. */
  const pageTotalForNotebook =
    (notebookMeta &&
      (notebookMeta.pageCount > 0
        ? notebookMeta.pageCount
        : notebookMeta.hasWritings
          ? 1
          : 0)) ||
    rmLinesEnv?.pageCount ||
    0;

  const useRmLines = rmLinesChecked && rmLinesEnv !== null;
  const showRmPager = useRmLines && pageTotalForNotebook > 0;

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
        <ButtonToolbar className="gap-2 w-100 align-items-center">
          {showRmPager && (
            <ButtonGroup aria-label="Notebook page">
              <Button
                size="sm"
                variant="outline-secondary"
                disabled={rmPage <= 1}
                onClick={() => setRmPage((p) => Math.max(1, p - 1))}
              >
                <FaChevronLeft />
              </Button>
              <Button
                size="sm"
                variant="outline-secondary"
                disabled={rmPage >= pageTotalForNotebook}
                onClick={() => setRmPage((p) => Math.min(pageTotalForNotebook, p + 1))}
              >
                <FaChevronRight />
              </Button>
            </ButtonGroup>
          )}
          {showRmPager && (
            <span className="text-muted small me-2">
              Page {rmPage} of {notebookPageCount}
            </span>
          )}
          <div className="flex-spacer" />
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
          {!rmLinesChecked && (
            <p style={{ padding: 16 }} className="text-muted">
              Loading notebook preview…
            </p>
          )}
          {rmLinesChecked && useRmLines && pageTotalForNotebook > 0 && (
            <RmLinesRenderer environment={rmLinesEnv} page={rmPage - 1} />
          )}
          {rmLinesChecked && useRmLines && pageTotalForNotebook <= 0 && notebookMeta && (
            <p style={{ padding: 16 }} className="text-muted">
              No pages to preview.
            </p>
          )}
          {rmLinesChecked && !useRmLines && notebookMeta === null && (
            <p style={{ padding: 16 }} className="text-muted">
              Loading notebook preview…
            </p>
          )}
          {rmLinesChecked && !useRmLines && notebookMeta && notebookPageCount > 0 && (
            Array.from({ length: notebookPageCount }, (_, i) => {
              const pageNum = i + 1;
              return (
                <NotebookPageSvg
                  key={pageNum}
                  docId={file.id}
                  pageNum={pageNum}
                  label={`${data.name} page ${pageNum}`}
                />
              );
            })
          )}
          {rmLinesChecked && !useRmLines && notebookMeta && notebookPageCount <= 0 && (
            <p style={{ padding: 16 }} className="text-muted">
              No pages to preview.
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
