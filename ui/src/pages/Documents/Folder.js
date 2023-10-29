import FileList from "./FileList";
import FileIcon from "./FileIcon";
import Navbar from 'react-bootstrap/Navbar';
//import apiservice from "../../services/api.service"

export default function Folder({ folder }) {
  return (
    <>
      <Navbar>
        { folder && (
          <h6>
            <FileIcon file={folder} />
            {folder.name}
          </h6>) }
      </Navbar>

      <div>
        <FileList files={folder.children} />
      </div>
    </>
  );
}
