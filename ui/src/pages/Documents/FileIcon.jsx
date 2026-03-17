import { BsFilePdf, BsFolder, BsFileEarmark, BsFileEarmarkText, BsCloud, BsFile, BsTrash, BsGrid3X3, BsBook } from "react-icons/bs";

export default function FileIcon({ file }) {

  const Icon = () => {
    if (!!file.icon) {
      switch (file.icon) {
        case "device":
          return <BsFile />
        case "trash":
          return <BsTrash />
        case "templates":
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
      return <BsFilePdf />
    }

    if (file.type === "epub" || (file.name && file.name.toLowerCase().endsWith(".epub"))) {
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
