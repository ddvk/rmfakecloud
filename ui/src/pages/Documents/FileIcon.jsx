import { useState } from "react";
import { BsFilePdf, BsFolder, BsFileEarmark, BsFileEarmarkText, BsCloud, BsFile, BsTrash, BsGrid3X3, BsBook } from "react-icons/bs";
import apiservice from "../../services/api.service";

export default function FileIcon({ file, showThumbnail = false }) {
  const [epubThumbFailed, setEpubThumbFailed] = useState(false);

  const Icon = () => {
    if (!!file.icon) {
      switch (file.icon) {
        case "device":
          return <BsFile />
        case "trash":
          return <BsTrash />
        case "templates":
        case "templates-methods":
          return <BsFileEarmarkText />
        case "methods":
          return <BsGrid3X3 />
        case "cloud":
          return <BsCloud />
        default:
          return <BsFile />
      }
    }

    if (file.isFolder) {
      return <BsFolder />
    }

    if (file.type === "pdf") {
      if (showThumbnail && file?.id) {
        return (
          <img
            src={apiservice.getDocumentPageBackgroundUrl(file.id, 1)}
            alt={file.name || "PDF"}
            style={{
              width: 68,
              height: 88,
              objectFit: "cover",
              borderRadius: 3,
              border: "1px solid #dee2e6",
              background: "#fff",
            }}
          />
        );
      }
      return <BsFilePdf />
    }

    if (file.type === "epub" || (file.name && file.name.toLowerCase().endsWith(".epub"))) {
      if (showThumbnail && file?.id && !epubThumbFailed) {
        return (
          <img
            src={apiservice.getEpubCoverThumbUrl(file.id)}
            alt={file.name || "EPUB"}
            onError={() => setEpubThumbFailed(true)}
            style={{
              width: 68,
              height: 88,
              objectFit: "cover",
              borderRadius: 3,
              border: "1px solid #dee2e6",
              background: "#fff",
            }}
          />
        );
      }
      return <BsBook />
    }

    if (file.type === "notebook") {
      return <BsFileEarmarkText />
    }

    if (file.type === "template") {
      return <BsFileEarmarkText />
    }

    if (file.type === "method") {
      return <BsGrid3X3 />
    }

    return <BsFileEarmark />
  }

  return (
    <span style={{ padding: '0 0.5em 0 0' }}>
      <Icon />
    </span>
  );
}
