import { React, useEffect, useState } from 'react';
import apiservice from "../../services/api.service"
import { Tree } from 'react-arborist';
import { MdArrowDropDown, MdArrowRight } from "react-icons/md";
import { BsSearch } from "react-icons/bs";
import Form from 'react-bootstrap/Form';
import InputGroup from 'react-bootstrap/InputGroup';
import Button from 'react-bootstrap/Button';
import FileIcon from './FileIcon';

const DocumentTree = ({ onFileSelected }) => {
  const [term, setTerm] = useState("");
  const [name, setName] = useState("");
  const [selected, setSelected] = useState(null);

  const onSelect = (node) => {
    setSelected(node);
    const file = node ? node.data : null
    onFileSelected(file);
  }

  const createFolder = async () => {
    let parentId = "";
    if (selected && selected.data.isFolder) {
      parentId = selected.id;
    }
    await apiservice.createFolder({ name, parentId});
    await loadDocs();
  }

  function FolderArrow({ node }: { node: NodeApi }) {
    if (node.isLeaf) {
      return <FileIcon file={node.data} />
    } else {
      return (<>
        {node.isOpen ? <MdArrowDropDown /> : <MdArrowRight />}
      </>);
    }
  }

  function Node({ node, style, dragHandle }: NodeRendererProps) {
    const isSelected = selected && node.data.id === selected.id;
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
        <h6 style={{ padding: '0 5px', marginTop: '0.5rem' }}>
          <FolderArrow node={node} />
          {node.data.name}
        </h6>
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
        <div>

          <InputGroup className="mb-3">
            <InputGroup.Text>
              <BsSearch />
            </InputGroup.Text>

            <Form.Control type="text" value={term} onChange={(e) => { setTerm(e.currentTarget.value); onSelect(null) }} />
          </InputGroup>
          {/*
      { dwn && <button onClick={onDownloadClick}>Download {dwn.name}</button> }
      { downloadError && <div class="error">{downloadError}</div> }
      <Treebeard style={treeStyle} data={data.docs} animations={false} onToggle={onToggle} />
      */}
          <Tree
            data={entries}
            rowHeight={36}
            indent={36}
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
            <InputGroup className="mb-3">
            <Form.Control type="text" value={name} onChange={(e) => setName(e.currentTarget.value)} />

              <Button onClick={createFolder}>Create</Button>
          </InputGroup>
        </div>
      )
}
export default DocumentTree;
