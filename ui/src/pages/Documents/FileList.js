import { Stack, Button } from "react-bootstrap";
import Navbar from 'react-bootstrap/Navbar';
import FileIcon from "./FileIcon";

//import apiservice from "../../services/api.service"

export default function FileListViewer({ files }) {

  const folderItems = files.map(file =>
    <div className="p-2" key={file.id}>
      <FileIcon file={file} />
      {file.name}
    </div>
  );

  return (
    <>
      <Navbar style={{ borderBottom: '1px solid #eee' }}>
        <Button size="sm">create Folder</Button>
        <Button size="sm">upload File</Button>
      </Navbar>

      {files && (
        <Stack gap={3}>{folderItems}</Stack>
      )}
    </>
  );
}
