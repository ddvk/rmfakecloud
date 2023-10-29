import { useState } from "react";
import { Stack, ToggleButton, Button, ToggleButtonGroup } from "react-bootstrap";
import Navbar from 'react-bootstrap/Navbar';
import FileIcon from "./FileIcon";
//import apiservice from "../../services/api.service"
import { BsFillGridFill } from "react-icons/bs";
import { FaList } from "react-icons/fa";

export default function FileListViewer({ files, onSelect }) {

  const [listStyle, setListStyle] = useState("grid");

  const onClickItem = (file) => {
    onSelect(file.id);
  }

  const onCreateFolderClick = () => {
    console.log('not yet implemented');
  }

  const onUploadFileClick = () => {
    console.log('not yet implemented');
  }

  const itemClassName = (item) => {
    if (item.isFolder) return "is-folder";
    return "";
  }

  const listItems = files.map(file =>
    <div className="filelist-item p-2" key={file.id} onClick={() => onClickItem(file)}>
      <div className={itemClassName(file)}>
        <FileIcon file={file} />
        {file.name}
      </div>
    </div>
  );

  const gridFolderItems = files.filter(file => file.isFolder).map(file =>
    <div className="filegrid-folder-item" key={file.id} onClick={() => onClickItem(file)}>
      <div>
        <FileIcon file={file} />
        {file.name}
      </div>
    </div>
  );

  const gridFileItems = files.filter(file => !file.isFolder).map(file =>
    <div className="filegrid-file-item" key={file.id} onClick={() => onClickItem(file)}>
      <div className="fileicon">
        <FileIcon file={file} />
      </div>
      <div className="filename">
        {file.name}
      </div>
    </div>
  );

  return (
    <>
      <Navbar style={{ borderBottom: '1px solid #eee' }}>
        <Button size="sm" variant="outline" onClick={onCreateFolderClick}>create Folder</Button>
        <Button size="sm" variant="outline" onClick={onUploadFileClick}>upload File</Button>
        <div style={{flex:1}}></div>
        <ToggleButtonGroup value={listStyle} onChange={(v) => setListStyle(v)} name="abc">
          <ToggleButton id="grid" name="grid" size="sm" value="grid" variant="outline">
            <BsFillGridFill />
          </ToggleButton>
          <ToggleButton id="list" name="list" size="sm" value="list" variant="outline">
            <FaList />
          </ToggleButton>
        </ToggleButtonGroup>
      </Navbar>

      {files && (listStyle === "list") && (
        <Stack className="filelist">{listItems}</Stack>
      )}

      {files && (listStyle === "grid") && (
        <div>
          <div className="filegrid">
            {gridFolderItems}
          </div>
          <div style={{ height: '1em', width: '100%' }}></div>
          <div className="filegrid">
            {gridFileItems}
          </div>
        </div>
      )}
    </>
  );
}
