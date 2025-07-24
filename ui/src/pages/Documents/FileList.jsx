import { Stack } from "react-bootstrap";
import FileIcon from "./FileIcon";
import moment from "moment";

// optional: load user-specific locale from config
// moment.locale(locale); // eg. 'de' or 'fr'

function formatBytes(bytes) {
  if(bytes === 0) return '0 Bytes';
  let k = 1024,
    sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'],
    i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

export default function FileListViewer({ listStyle, files, onSelect, counter, selectedIds = [], onSelectItem }) {
  const onClickItem = (file) => {
    onSelect(file);
  }

  const isFolderClassName = (item) => {
    if (item.isFolder) return "is-folder";
    return "";
  }

  const listItems = files.map(file =>
    <div className="filelist-item p-2" key={file.id} onClick={() => onClickItem(file)}>
      <Stack direction="horizontal">
        <div>
          <input
            type="checkbox"
            checked={selectedIds.includes(file.id)}
            onClick={e => e.stopPropagation()}
            onChange={() => onSelectItem && onSelectItem(file.id)}
            className="filelist-checkbox"
          />
        </div>
        <div>
          <FileIcon file={file.data} />
        </div>
        <div className={isFolderClassName(file)}>
          {file.data.name}
        </div>
        <div style={{ flex: 1 }}>
        </div>
        <div className="filelist-metadata">
          {!file.isLeaf && <span>{file.children.length || "empty"}</span> }
          {file.isLeaf && <span>{formatBytes(file.data.size)}</span> }
        </div>
        <div className="filelist-metadata">
          {moment(file.data.lastModified).format('L')}{" "}
          {moment(file.data.lastModified).format('LT')}
        </div>
      </Stack>
    </div>
  );

  const gridFolderItems = files.filter(file => !file.isLeaf).map(file =>
    <div className="filegrid-folder-item" key={file.id} onClick={() => onClickItem(file)}>
      <div>
        <FileIcon file={file.data} />
        {file.data.name}
      </div>
    </div>
  );

  const gridFileItems = files.filter(file => file.isLeaf).map(file =>
    <div className="filegrid-file-item" key={file.id} onClick={() => onClickItem(file)}>
      <div className="fileicon">
        <FileIcon file={file.data} />
      </div>
      <div className="filename">
        {file.data.name}
      </div>
      <div className="filegrid-metadata">
        <span>{formatBytes(file.data.size)}</span>
      </div>
    </div>
  );

  return (
    <>
      {files && (listStyle === "list") && (
        <div>
          <div style={{ height: '1em', width: '100%' }}></div>
          <Stack className="filelist">{listItems}</Stack>
        </div>
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
