import { React, useEffect, useState, useRef } from 'react';
import apiservice from "../../services/api.service"
import { Tree } from 'react-arborist';
//import { AiOutlinePlusSquare, AiOutlineMinusSquare } from "react-icons/ai";
//import Button from 'react-bootstrap/Button';
import FileIcon from './FileIcon';

const DocumentTree = ({ onNodeSelected, term }) => {
  //const [name, setName] = useState("");
  const treeRef = useRef();

  const onSelect = (selection) => {
    if (selection.length > 0) {
      const node = selection[0];
      onNodeSelected(node);
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
        onClick={() => node.isInternal && node.toggle()}
      >
        <h6 style={{ padding: '0 5px', marginTop: '0.5rem' }}>
          {/* TODO: decide on how to make this look not ugly
          <span onClick={() => node.isInternal && node.toggle()}>
            <FolderArrow node={node} />
          </span>
          */}
          <FileIcon file={node.data} />
          {node.data.name}
        </h6>
      </div>
    );
  }

  function Cursor({ top, left }: CursorProps) {
    return <div style={{ top, left }}></div>;
  }

  const [entries, setEntries] = useState([]);

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState();

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
    // pass data to tree lib
    setEntries([root, trash]);
    // autoSelect root node
    //onSelect({ data: root });

    const tree = treeRef.current;
    tree.open("root");
    tree.get("root").select();
  }

  const onCreate = ({ parentId, index, type }) => {};
  const onRename = ({ id, name }) => {};
  const onMove = ({ dragIds, parentId, index }) => {};
  const onDelete = ({ ids }) => {};

  useEffect(() => {
    loadDocs()

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
        ref={treeRef}
        data={entries}
        rowHeight={36}
        indent={36}
        width="100%"
        height={700}
        renderCursor={Cursor}
        searchTerm={term}
        onCreate={onCreate}
        onRename={onRename}
        onSelect={onSelect}
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
