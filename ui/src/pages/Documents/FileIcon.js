import { BsFiletypePdf } from "react-icons/bs";

export default function FileIcon({ file }) {

  const Icon = BsFiletypePdf;

  return (
    <span style={{ padding: '0 0.5em 0 0' }}>
      <Icon />
    </span>
  );
}
