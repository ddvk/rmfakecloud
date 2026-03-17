import Navbar from "react-bootstrap/Navbar";
import NameTag from "../../components/NameTag";
import constants from "../../common/constants";

/**
 * Renders an EPUB as a native website: backend serves unpacked HTML/CSS/assets by path;
 * the iframe loads the first spine item and relative links work like a normal site.
 */
export default function EpubViewer({ file, onSelect }) {
  const { data } = file;
  const epubBase =
    typeof window !== "undefined"
      ? `${window.location.origin}${constants.ROOT_URL}/documents/${data?.id}/epub/`
      : "";

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
        {data?.id && (
          <iframe
            title={file?.data?.name || "EPUB"}
            src={epubBase}
            style={{
              flex: 1,
              minHeight: 0,
              width: "100%",
              border: "none",
            }}
          />
        )}
      </div>
    </>
  );
}
