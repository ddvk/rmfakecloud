import { useState } from "react";
import Navbar from 'react-bootstrap/Navbar';
import { Button, InputGroup, Form } from "react-bootstrap";
import { BsChevronRight } from "react-icons/bs";
import Modal from 'react-bootstrap/Modal';
import { BsFillGridFill } from "react-icons/bs";
import { FaList } from "react-icons/fa";

import { ToggleButton, ToggleButtonGroup } from "react-bootstrap";

import FileList from "./FileList";
import apiservice from "../../services/api.service"

import Upload from "./Upload"
import styles from "./Documents.module.scss"

export default function Folder({ folder, onSelect, onUpdate }) {
  const [listStyle, setListStyle] = useState("grid");
  const [folderName, setFolderName] = useState("");
  const [showCreateFileModal, setShowCreateFolder] = useState(false);

  const { data, id } = folder;

  const onCreateFolderClick = async () => {
    const res = await apiservice.createFolder({ name: folderName, parentId: id});
    console.log("created folder with id", res.ID);
    setFolderName("");
    setShowCreateFolder(false);

    onUpdate();
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

  const fileUploaded = () => {
    onUpdate();
  }

  return (
    <>
      <Navbar style={{ marginLeft: '-12px' }}>
        { folder && (<div><NameTag node={folder} /></div>) }
      </Navbar>

      <Navbar className={styles.filedivider}>
        <Button size="sm" variant="outline" onClick={() => setShowCreateFolder(true)}>Create Folder</Button>
        <div className={styles.stretch}></div>
        <ToggleButtonGroup value={listStyle} onChange={(v) => setListStyle(v)} name="abc">
          <ToggleButton id="grid" name="grid" size="sm" value="grid" variant="outline">
            <BsFillGridFill />
          </ToggleButton>
          <ToggleButton id="list" name="list" size="sm" value="list" variant="outline">
            <FaList />
          </ToggleButton>
        </ToggleButtonGroup>
      </Navbar>

      <Upload filesUploaded={fileUploaded} uploadFolder={id}></Upload>
      <FileList listStyle={listStyle} files={data.children} onSelect={onSelect} />

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
