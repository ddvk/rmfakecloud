import { useCallback, useMemo, useState } from "react";
import Navbar from 'react-bootstrap/Navbar';
import { Button, InputGroup, Form } from "react-bootstrap";
import { BsChevronRight } from "react-icons/bs";
import { useDropzone } from 'react-dropzone';
import Modal from 'react-bootstrap/Modal';
import { BsFillGridFill } from "react-icons/bs";
import { FaList } from "react-icons/fa";

import { ToggleButton, ToggleButtonGroup } from "react-bootstrap";

import FileList from "./FileList";
import apiservice from "../../services/api.service"


const baseStyle = {
  transition: 'background 0.3s ease',
  backgroundColor: '#000'
};
const acceptStyle = {
  backgroundColor: 'rgb(13, 110, 253, 0.3)'
};
const rejectStyle = {
  backgroundColor: 'rgba(255, 255, 255, 0.3)'
};

export default function Folder({ folder, onSelect }) {
  const [listStyle, setListStyle] = useState("grid");
  const [folderName, setFolderName] = useState("");
  const [showCreateFileModal, setShowCreateFolder] = useState(false);

  const { data, id } = folder;
  const uploadFolder = id;

  const onCreateFolderClick = async () => {
    const res = await apiservice.createFolder({ name: folderName, parentId: id});
    console.log("created folder with id", res.ID);
    setFolderName("");
    setShowCreateFolder(false);

    //TODO: update
  }

  const onUploadFileClick = () => {
    console.log('not yet implemented');
  }

  const NameTag = ({ node }) => {
    if (node.parent) {
      return (<>
        <NameTag node={node.parent} />
        { !node.parent.isRoot && <BsChevronRight /> }
        <Button variant="outline" onClick={() => onSelect(node.id)}>{node.data.name}</Button>
      </>)
    }
    return <></>
  }

  const onReject = () => {
    console.log('drop rejected');
  }

  const onUploadDone = useCallback((res) => {
    const upload = res.pop();
    console.log("finished uploading file", upload);

    //TODO: update
  }, []);

  const onDrop = useCallback(async (acceptedFiles) => {
    try {
      //setUploading(true);
      const res = await apiservice.upload(uploadFolder, acceptedFiles)
      onUploadDone(res);
      //setLastError(null)
      //props.filesUploaded()
    } catch (e) {
      //setLastError(e)
      console.error(e)
    } finally{
      //setUploading(false);
    }
    console.log('done')
  }, [uploadFolder, onUploadDone]);

  const {
    getRootProps,
    getInputProps,
    isDragAccept,
    isDragReject
  } = useDropzone({
    accept: 'application/pdf, application/zip, application/epub+zip',
    onDropAccepted: onDrop,
    onDropRejected: onReject,
    noClick: true,
    noKeyboard: true,
    multiple: true,
    noDragEventsBubbling: true,
  });

  const style = useMemo(() => ({
    ...baseStyle,
    ...(isDragAccept ? acceptStyle : {}),
    ...(isDragReject ? rejectStyle : {})
  }), [
    isDragReject,
    isDragAccept
  ]);

  return (
    <>
      <Navbar style={{ marginLeft: '-12px' }}>
        { folder && (<div><NameTag node={folder} /></div>) }
      </Navbar>

      <Navbar style={{ borderBottom: '1px solid #eee' }}>
        <Button size="sm" variant="outline" onClick={() => setShowCreateFolder(true)}>create Folder</Button>
        <Button size="sm" variant="outline" onClick={onUploadFileClick}>upload File</Button>
        <div style={{flex:1}}></div>
        <ToggleButtonGroup value={listStyle} onChange={(v) => setListStyle(v)} name="abc">
          <ToggleButton id="grid" name="grid" size="sm" value="grid" variant="outline">
            <BsFillGridFill />
          </ToggleButton>
          <ToggleButton id="list" name="list" size="sm" value="list" variant="outline">
            <FaList />
          </ToggleButton>
        </ToggleButtonGroup>
      </Navbar>

      <div {...getRootProps({ style })}>
        <input {...getInputProps()} />
        <FileList listStyle={listStyle} files={data.children} onSelect={onSelect} />
      </div>

      <Modal show={showCreateFileModal} onHide={() => setShowCreateFolder(false)}>
        <Modal.Header closeButton>
          Create a new folder
        </Modal.Header>

        <Modal.Body>
          <InputGroup className="mb-3">
            <Form.Control type="text" value={folderName} onChange={(e) => setFolderName(e.currentTarget.value)} />

            <Button onClick={onCreateFolderClick}>Create</Button>
          </InputGroup>
        </Modal.Body>

      </Modal>
    </>
  );
}
