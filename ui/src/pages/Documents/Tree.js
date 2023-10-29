import { React, useEffect, useState } from 'react';
import apiservice from "../../services/api.service"
import { Tree } from 'react-arborist';
//import { AiOutlinePlusSquare, AiOutlineMinusSquare } from "react-icons/ai";
//import Button from 'react-bootstrap/Button';
import FileIcon from './FileIcon';

const DocumentTree = ({ selection, onSelect, term }) => {
  const onTreeSelect = (sel) => {
    if (sel.length > 0) {
      const node = sel[0];
      node.open();
      onSelect(node);
    }
  }

  /*
  const createFolder = async () => {
    let parentId = "";
    if (selected && selected.data.isFolder) {
      parentId = selected.id;
    }
    await apiservice.createFolder({ name, parentId});
    await loadDocs();
  }

  function FolderArrow({ node }: { node: NodeApi }) {
    if (node.isLeaf) return <></>;
    return (<>
      {node.isOpen ? <AiOutlineMinusSquare /> : <AiOutlinePlusSquare />}
    </>);
  }
  */

  function Node({ node, style, dragHandle }: NodeRendererProps) {
    const selectedStyle = { ...style, color: '#0d6efd' };
    return (
      <div
        style={ node.isSelected ? selectedStyle : style}
        ref={dragHandle}
        onClick={() => node.isInternal && node.open()}
      >
        <div className={itemClassName(node.data)}>
          {/* TODO: decide on how to make this look not ugly
          <span onClick={() => node.isInternal && node.toggle()}>
            <FolderArrow node={node} />
          </span>
          */}
          <FileIcon file={node.data} />
          {node.data.name}
        </div>
      </div>
    );
  }

  function Cursor({ top, left }: CursorProps) {
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
      name: "reMarkable",
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
      {/*
      <InputGroup className="mb-3">
        <Form.Control type="text" value={name} onChange={(e) => setName(e.currentTarget.value)} />

        <Button onClick={createFolder}>Create</Button>
      </InputGroup>
      */}
    </div>
  )
}
export default DocumentTree;
