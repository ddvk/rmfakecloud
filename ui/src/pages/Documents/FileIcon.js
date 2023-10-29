import { BsFilePdf, BsFolder, BsFileEarmark, BsFileEarmarkText, BsCloud, BsFile, BsTrash } from "react-icons/bs";

export default function FileIcon({ file }) {

  const Icon = () => {
    if (!!file.icon) {
      switch (file.icon) {
        case "device":
          return <BsFile />
        case "trash":
          return <BsTrash />
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

    if (file.type === "notebook") {
      return <BsFileEarmarkText />
    }

    return <BsFileEarmark />
  }

  return (
    <span style={{ padding: '0 0.5em 0 0' }}>
      <Icon />
    </span>
  );
}
