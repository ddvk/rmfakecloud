import { useState, useEffect, useRef } from "react";
import Navbar from "react-bootstrap/Navbar";
import apiservice from "../../services/api.service";
import NameTag from "../../components/NameTag";

/**
 * Renders an EPUB document like a website: continuous scrolling, HTML content in a doc-like view.
 */
export default function EpubViewer({ file, onSelect }) {
  const { data } = file;
  const containerRef = useRef(null);
  const bookRef = useRef(null);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!data?.id || !containerRef.current) return;

    let objectUrl = null;

    const initEpub = (url) => {
      return import("epubjs").then(({ default: ePub }) => {
        const book = ePub(url, {
          requestCredentials: true,
        });
        bookRef.current = book;
        return book.opened;
      });
    };

    // Fetch with credentials and use blob URL so epubjs can load it (avoids CORS/credentials in its request)
    apiservice
      .download(data.id, "epub")
      .then((blob) => {
        objectUrl = URL.createObjectURL(blob);
        return initEpub(objectUrl);
      })
      .then(() => {
        const book = bookRef.current;
        if (!book || !containerRef.current) return;
        const rendition = book.renderTo(containerRef.current, {
          manager: "continuous",
          flow: "scrolled-doc",
          width: "100%",
          height: "100%",
        });
        return rendition.display();
      })
      .then(() => {
        setLoading(false);
        setError(null);
      })
      .catch((err) => {
        setError(err?.message || "Failed to load EPUB");
        setLoading(false);
      });

    return () => {
      if (objectUrl) URL.revokeObjectURL(objectUrl);
      if (bookRef.current?.rendition) {
        try {
          bookRef.current.rendition.destroy();
        } catch (_) {}
      }
      bookRef.current = null;
    };
  }, [data.id]);

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
        style={{
          flex: 1,
          minHeight: 0,
          display: "flex",
          flexDirection: "column",
          overflow: "hidden",
          background: "#fff",
        }}
      >
        {loading && (
          <div className="p-3 text-muted text-center">Loading…</div>
        )}
        {error && (
          <div className="p-3 text-danger text-center">{error}</div>
        )}
        <div
          ref={containerRef}
          style={{
            flex: 1,
            minHeight: 0,
            overflow: "auto",
          }}
        />
      </div>
    </>
  );
}
