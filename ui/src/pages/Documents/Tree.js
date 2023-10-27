import { React, useEffect, useState } from 'react';
import apiservice from "../../services/api.service"
import { Tree } from 'react-arborist';
import { MdArrowDropDown, MdArrowRight } from "react-icons/md";
import { BsTree } from "react-icons/bs";

const DocumentTree = ({ onFileSelected }) => {

const onSelect = (node) => {
  onFileSelected(node);
}

function Input({ node }: { node: NodeApi<GmailItem> }) {
  return (
    <input
      autoFocus
      type="text"
      defaultValue={node.data.name}
      onFocus={(e) => e.currentTarget.select()}
      onBlur={() => node.reset()}
      onKeyDown={(e) => {
        if (e.key === "Escape") node.reset();
        if (e.key === "Enter") node.submit(e.currentTarget.value);
      }}
    />
  );
}

function FolderArrow({ node }: { node: NodeApi<GmailItem> }) {
  if (node.isLeaf) return <span></span>;
  return (
    <span>
      {node.isOpen ? <MdArrowDropDown /> : <MdArrowRight />}
    </span>
  );
}

function Node({ node, style, dragHandle }: NodeRendererProps<GmailItem>) {
  const Icon = node.data.icon || BsTree;
  return (
    <div
      ref={dragHandle}
      onClick={() => { onSelect(node); return node.isInternal && node.toggle(); }}
    >
      <FolderArrow node={node} />
      <span>
        <Icon />
      </span>
      <span>{node.isEditing ? <Input node={node} /> : node.data.name}</span>
      <span>{node.data.unread === 0 ? null : node.data.unread}</span>
    </div>
  );
}

  const [entries, setEntries] = useState([]);
  const [trash, setTrash] = useState([]);

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
    setEntries(Entries);
    setTrash(Trash);
    console.log("Trash:", trash);
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
  /*
  const onToggle = (node, toggled) => {
    if (cursor) {
      cursor.active = false;
    }
    node.active = true;
    console.log(node.id)
    if (node.children) {
      node.toggled = toggled;
      setDownloadUrl(null);
      if (onFileSelected) {
        onFileSelected(null);
      }
      props.onFolderChanged(node.id);
    } else {
      //TODO: another quick poc hack
      setDownloadUrl({id:node.id, name:node.name})
      if (onFileSelected) {
        onFileSelected(node.id);
      }
    }
    setCursor(node);
    setData(Object.assign({}, data))
  }

  */

      return (
        <div style={{"marginTop":"20px"}}>
          {/*
      { dwn && <button onClick={onDownloadClick}>Download {dwn.name}</button> }
      { downloadError && <div class="error">{downloadError}</div> }
      <Treebeard style={treeStyle} data={data.docs} animations={false} onToggle={onToggle} />
      */}
          <Tree
            data={entries}
            onCreate={onCreate}
            onRename={onRename}
            onMove={onMove}
            onDelete={onDelete}
          >
            {Node}
          </Tree>
        </div>
      )
}
export default DocumentTree;
