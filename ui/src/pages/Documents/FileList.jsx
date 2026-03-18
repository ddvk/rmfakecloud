import { useEffect, useMemo, useRef, useState } from "react";
import { Stack } from "react-bootstrap";
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  flexRender,
} from "@tanstack/react-table";
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

  const [contextMenu, setContextMenu] = useState(null); // { x, y, file }
  const contextMenuRef = useRef(null);

  useEffect(() => {
    if (!contextMenu) return;
    const onDown = (e) => {
      // Close if click outside the menu.
      if (contextMenuRef.current && !contextMenuRef.current.contains(e.target)) {
        setContextMenu(null);
      }
    };
    const onKey = (e) => {
      if (e.key === "Escape") setContextMenu(null);
    };
    window.addEventListener("mousedown", onDown);
    window.addEventListener("keydown", onKey);
    return () => {
      window.removeEventListener("mousedown", onDown);
      window.removeEventListener("keydown", onKey);
    };
  }, [contextMenu]);

  const openContextMenu = (e, file) => {
    e.preventDefault();
    e.stopPropagation();
    const pad = 8;
    const vw = window.innerWidth || 0;
    const vh = window.innerHeight || 0;
    const menuW = 240;
    const menuH = 210;
    const x = Math.min(Math.max(pad, e.clientX), Math.max(pad, vw - menuW - pad));
    const y = Math.min(Math.max(pad, e.clientY), Math.max(pad, vh - menuH - pad));
    setContextMenu({ x, y, file });
  };

  const contextHeader = useMemo(() => {
    if (!contextMenu?.file) return "";
    if (selectedIds.length > 1) return `${selectedIds.length} selected`;
    return contextMenu.file?.data?.name || contextMenu.file?.id || "Actions";
  }, [contextMenu?.file, selectedIds.length]);

  const onActionOpen = () => {
    if (!contextMenu?.file) return;
    onClickItem(contextMenu.file);
    setContextMenu(null);
  };

  const isFolderClassName = (item) => {
    if (item.isFolder) return "is-folder";
    return "";
  }

  // Define columns for TanStack Table
  const columns = useMemo(
    () => [
      {
        id: 'select',
        header: ({ table }) => (
          <input
            type="checkbox"
            checked={table.getIsAllRowsSelected()}
            indeterminate={table.getIsSomeRowsSelected()}
            onChange={table.getToggleAllRowsSelectedHandler()}
            onClick={(e) => e.stopPropagation()}
          />
        ),
        cell: ({ row }) => (
          <input
            type="checkbox"
            checked={selectedIds.includes(row.original.id)}
            onChange={() => onSelectItem && onSelectItem(row.original.id)}
            onClick={(e) => e.stopPropagation()}
            className="filelist-checkbox"
          />
        ),
        enableSorting: false,
        size: 40,
      },
      {
        id: 'icon',
        header: '',
        cell: ({ row }) => <FileIcon file={row.original.data} />,
        enableSorting: false,
        size: 40,
      },
      {
        accessorKey: 'data.name',
        id: 'name',
        header: 'Name',
        cell: ({ row }) => (
          <div className={isFolderClassName(row.original)}>
            {row.original.data.name}
          </div>
        ),
        sortingFn: (rowA, rowB) => {
          return rowA.original.data.name.localeCompare(rowB.original.data.name);
        },
      },
      {
        id: 'metadata',
        header: 'Size',
        accessorFn: (row) => row.isLeaf ? row.data.size : row.children?.length || 0,
        cell: ({ row }) => (
          <div className="filelist-metadata">
            {!row.original.isLeaf && <span>{row.original.children?.length || "empty"}</span>}
            {row.original.isLeaf && <span>{formatBytes(row.original.data.size)}</span>}
          </div>
        ),
        sortingFn: (rowA, rowB) => {
          const valA = rowA.original.isLeaf ? rowA.original.data.size : rowA.original.children?.length || 0;
          const valB = rowB.original.isLeaf ? rowB.original.data.size : rowB.original.children?.length || 0;
          return valA - valB;
        },
      },
      {
        accessorKey: 'data.lastModified',
        id: 'lastModified',
        header: 'Last Modified',
        cell: ({ row }) => (
          <div className="filelist-metadata">
            {moment(row.original.data.lastModified).format('L')}{" "}
            {moment(row.original.data.lastModified).format('LT')}
          </div>
        ),
        sortingFn: (rowA, rowB) => {
          const dateA = new Date(rowA.original.data.lastModified);
          const dateB = new Date(rowB.original.data.lastModified);
          return dateA.getTime() - dateB.getTime();
        },
        sortDescFirst: true, // Default to newest first for this column
      },
    ],
    [selectedIds, onSelectItem]
  );

  const table = useReactTable({
    data: files,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    enableSortingRemoval: false, // Always have a sort direction
    initialState: {
      sorting: [
        {
          id: 'lastModified',
          desc: true, // Default: newest first
        },
      ],
    },
  });

  const gridFolderItems = files.filter(file => !file.isLeaf).map(file =>
    <div className="filegrid-folder-item" key={file.id} onContextMenu={(e) => openContextMenu(e, file)}>
      <input
        type="checkbox"
        checked={selectedIds.includes(file.id)}
        onClick={e => e.stopPropagation()}
        onChange={() => onSelectItem && onSelectItem(file.id)}
        className="filelist-checkbox"
      />
      <div className="filegrid-checkbox-spacer" onClick={() => onClickItem(file)}>
        <FileIcon file={file.data} />
        {file.data.name}
      </div>
    </div>
  );

  const gridFileItems = files.filter(file => file.isLeaf).map(file =>
    <div className="filegrid-file-item" key={file.id} onContextMenu={(e) => openContextMenu(e, file)}>
      <input
        type="checkbox"
        checked={selectedIds.includes(file.id)}
        onClick={e => e.stopPropagation()}
        onChange={() => onSelectItem && onSelectItem(file.id)}
        className="filelist-checkbox"
      />
      <div className="filegrid-checkbox-spacer fileicon" onClick={() => onClickItem(file)}>
        <FileIcon file={file.data} showThumbnail={true} />
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
      {contextMenu && (
        <div
          ref={contextMenuRef}
          style={{
            position: "fixed",
            left: contextMenu.x,
            top: contextMenu.y,
            width: 240,
            background: "#fff",
            border: "1px solid rgba(0,0,0,0.15)",
            borderRadius: 8,
            boxShadow: "0 10px 30px rgba(0,0,0,0.18)",
            zIndex: 5000,
            overflow: "hidden",
          }}
          onContextMenu={(e) => {
            e.preventDefault();
            e.stopPropagation();
          }}
        >
          <div
            style={{
              padding: "8px 10px",
              fontSize: 12,
              fontWeight: 700,
              color: "#212529",
              background: "#f8f9fa",
              borderBottom: "1px solid rgba(0,0,0,0.08)",
              whiteSpace: "nowrap",
              overflow: "hidden",
              textOverflow: "ellipsis",
            }}
            title={contextHeader}
          >
            {contextHeader}
          </div>
          <button
            type="button"
            style={menuItemStyle}
            onClick={onActionOpen}
          >
            Open
          </button>
          <button type="button" style={menuItemStyle} disabled>
            Download
          </button>
          <button type="button" style={menuItemStyle} disabled>
            Rename
          </button>
          <div style={{ height: 1, background: "rgba(0,0,0,0.08)" }} />
          <button type="button" style={{ ...menuItemStyle, color: "#dc3545" }} disabled>
            Delete
          </button>
        </div>
      )}
      {files && (listStyle === "list") && (
        <div>
          <div style={{ height: '1em', width: '100%' }}></div>
          <div className="filelist">
            {/* Table header */}
            {table.getHeaderGroups().map(headerGroup => (
              <div key={headerGroup.id} className="filelist-header">
                {headerGroup.headers.map(header => (
                  <div
                    key={header.id}
                    className={`filelist-header-cell ${header.column.getCanSort() ? 'sortable' : ''}`}
                    style={{
                      width: header.getSize() !== 150 ? `${header.getSize()}px` : 'auto',
                      flex: header.getSize() === 150 ? 1 : undefined,
                      cursor: header.column.getCanSort() ? 'pointer' : 'default',
                    }}
                    onClick={header.column.getToggleSortingHandler()}
                  >
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                    {header.column.getCanSort() && (
                      <span className="sort-indicator">
                        {header.column.getIsSorted() === 'asc' ? ' ↑' : header.column.getIsSorted() === 'desc' ? ' ↓' : ' ⇅'}
                      </span>
                    )}
                  </div>
                ))}
              </div>
            ))}

            {/* Table body */}
            <Stack gap={1}>
              {table.getRowModel().rows.map(row => (
                <div
                  key={row.id}
                  className="filelist-item p-2"
                  onClick={() => onClickItem(row.original)}
                  onContextMenu={(e) => openContextMenu(e, row.original)}
                >
                  <Stack direction="horizontal">
                    {row.getVisibleCells().map(cell => (
                      <div
                        key={cell.id}
                        style={{
                          width: cell.column.getSize() !== 150 ? `${cell.column.getSize()}px` : 'auto',
                          flex: cell.column.getSize() === 150 ? 1 : undefined,
                        }}
                      >
                        {flexRender(cell.column.columnDef.cell, cell.getContext())}
                      </div>
                    ))}
                  </Stack>
                </div>
              ))}
            </Stack>
          </div>
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

const menuItemStyle = {
  width: "100%",
  textAlign: "left",
  padding: "9px 10px",
  background: "transparent",
  border: "none",
  fontSize: 13,
  color: "#212529",
  cursor: "pointer",
};
