import { React, useEffect, useState } from 'react';
import apiservice from "../../services/api.service"
import { Tree } from 'react-arborist';
import { MdArrowDropDown, MdArrowRight } from "react-icons/md";

const DocumentTree = ({ onFileSelected }) => {
  const [term, setTerm] = useState("");
  const [selected, setSelected] = useState(null);

  const onSelect = (node) => {
    onFileSelected(node.data);
    setSelected(node);
  }

  function FolderArrow({ node }: { node: NodeApi }) {
    if (node.isLeaf) return <span></span>;
    return (
      <span style={{ width: '20px', marginRight: '5px' }}>
        {node.isOpen ? <MdArrowDropDown /> : <MdArrowRight />}
      </span>
    );
  }

  function Node({ node, style, dragHandle }: NodeRendererProps) {
    const isSelected = selected && node.data.id === selected.id;
    style = { ...style, paddingTop: '5px', paddingBottom: '5px' }
    const selectedStyle = { ...style, color: '#0d6efd' };
    return (
      <div
        style={ isSelected ? selectedStyle : style}
        ref={dragHandle}
        onClick={() => {
          node.toggle();
          onSelect(node);
        }}
      >
        <FolderArrow node={node} />
        <span>{node.data.name}</span>
      </div>
    );
  }

  function Cursor({ top, left }: CursorProps) {
    return <div style={{ top, left }}></div>;
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
          <input
            value={term}
            onChange={(e) => setTerm(e.currentTarget.value)}
          />
          {/*
      { dwn && <button onClick={onDownloadClick}>Download {dwn.name}</button> }
      { downloadError && <div class="error">{downloadError}</div> }
      <Treebeard style={treeStyle} data={data.docs} animations={false} onToggle={onToggle} />
      */}
          <Tree
            data={entries}
            rowHeight={32}
            indent={32}
            width="100%"
            renderCursor={Cursor}
            searchTerm={term}
            onCreate={onCreate}
            onRename={onRename}
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
