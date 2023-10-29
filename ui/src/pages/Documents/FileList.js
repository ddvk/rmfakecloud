import { Stack } from "react-bootstrap";
import FileIcon from "./FileIcon";

export default function FileListViewer({ listStyle, files, onSelect }) {

  const onClickItem = (file) => {
    onSelect(file.id);
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
      {files && (listStyle === "list") && (
        <Stack className="filelist">{listItems}</Stack>
      )}

      {files && (listStyle === "grid") && (
        <div>
          <div style={{ height: '1em', width: '100%' }}></div>
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
