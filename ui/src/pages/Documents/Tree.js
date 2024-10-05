import { React, useEffect, useState } from 'react';
import apiservice from "../../services/api.service"
import { Tree } from 'react-arborist';
import FileIcon from './FileIcon';

import styles from "./Documents.module.scss"

const DocumentTree = ({ selection, onSelect, term, counter }) => {

  const onTreeSelect = (sel) => {
    if (sel.length > 0) {
      const node = sel[0];
      onSelect(node);
    }
  }

  function Node({ node, style, dragHandle }) {
    return (
      <div
        style={style}
        ref={dragHandle}
        className={ node.isSelected ? styles.selected : ""}
      >
        <div className={itemClassName(node.data)}>
          <FileIcon file={node.data} />
          {node.data.name}
        </div>
      </div>
    );
  }

  function Cursor({ top, left }) {
    return <div style={{ top, left }}></div>;
  }

  const [entries, setEntries] = useState([]);

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState();

  const itemClassName = (item) => {
    if (item.isFolder) return "treeitem-nodename is-folder";
    return "treeitem-nodename";
  }

  const loadDocs = async () => {
    setLoading(true)
    const { Trash, Entries } = await apiservice.listDocument()
      .catch(e => {
        setLoading(false)
        setError(e)
      })
    setLoading(false)
    // create virtual root node
    const root = {
      id: "root",
      name: "My Files",
      isFolder: true,
      icon: "device",
      children: Entries,
    }
    const trash = {
      id: "trash",
      name: "Trash",
      isFolder: true,
      icon: "trash",
      children: Trash,
    }
    setEntries([root, trash]);
  }

  const onCreate = ({ parentId, index, type }) => {};
  const onRename = ({ id, name }) => {};
  const onMove = ({ dragIds, parentId, index }) => {};
  const onDelete = ({ ids }) => {};

  useEffect(() => {
    loadDocs();
  },[counter])

  useEffect(() => {
    loadDocs();

    // eslint-disable-next-line
  },[])

  if (loading) {
    return <div>Loading...</div>;
  }
  if (error) {
    return <div>{error}</div>;
  }
  if (entries && !entries.length) {
    return <div>No documents</div>;
  }
  return (
    <div>

      <Tree
        ref={(tree) => {
          global.tree = tree;
        }}
        data={entries}
        selection={selection}
        rowHeight={36}
        indent={36}
        width="100%"
        height={700}
        renderCursor={Cursor}
        searchTerm={term}
        onCreate={onCreate}
        onRename={onRename}
        onSelect={onTreeSelect}
        onMove={onMove}
        onDelete={onDelete}
        className="documents-tree"
        disableEdit={true}
        disableDrag={true}
        disableDrop={true}
        openByDefault={false}
      >
        {Node}
      </Tree>
    </div>
  )
}
export default DocumentTree;
