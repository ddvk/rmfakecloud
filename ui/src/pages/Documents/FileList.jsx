import { useMemo } from "react";
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
    <div className="filegrid-folder-item" key={file.id}>
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
    <div className="filegrid-file-item" key={file.id}>
      <input
        type="checkbox"
        checked={selectedIds.includes(file.id)}
        onClick={e => e.stopPropagation()}
        onChange={() => onSelectItem && onSelectItem(file.id)}
        className="filelist-checkbox"
      />
      <div className="filegrid-checkbox-spacer fileicon" onClick={() => onClickItem(file)}>
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
