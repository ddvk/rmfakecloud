import { useState } from "react";
import Navbar from 'react-bootstrap/Navbar';
import { Button, InputGroup, Form } from "react-bootstrap";
import Modal from 'react-bootstrap/Modal';
import { BsFillGridFill } from "react-icons/bs";
import { FaList } from "react-icons/fa";
import { ToggleButton, ToggleButtonGroup } from "react-bootstrap";

import apiservice from "../../services/api.service"
import styles from "./Documents.module.scss"

import Upload from "./Upload"
import FileList from "./FileList";
import NameTag from "../../components/NameTag"
import { toast } from "react-toastify";

// lists the folders of the tree with their depth, skipping the excluded ids and everything below them
const folderOptions = (entries, excluded, depth = 0) =>
  (entries || []).filter(e => e.isFolder && !excluded.includes(e.id)).flatMap(e => [
    { id: e.id, name: e.name, depth },
    ...folderOptions(e.children, excluded, depth + 1),
  ]);

export default function Folder({ selection, onSelect, onUpdate, entries }) {
  const [listStyle, setListStyle] = useState("list");
  const [folderName, setFolderName] = useState("");
  const [showCreateFileModal, setShowCreateFolder] = useState(false);
  const [selectedIds, setSelectedIds] = useState([]);
  const [showRenameModal, setShowRename] = useState(false);
  const [newName, setNewName] = useState("");
  const [showMoveModal, setShowMove] = useState(false);
  const [targetId, setTargetId] = useState("");

  const folder = selection

  const onCreateFolderClick = async () => {
    const res = await apiservice.createFolder({ name: folderName, parentId: selection.id });
    console.log("created folder with id", res.ID);
    setFolderName("");
    setShowCreateFolder(false);

    onUpdate();
  }

  const onDeleteClick = async () => {
    if (selectedIds.length === 0) return;
    if (!window.confirm(`Are you sure you want to delete the selected item(s)?`)) return;
    for (const id of selectedIds) {
      const file = folder.children.find(f => f.id === id);
      const name = file?.data?.name || id;
      try {
        await apiservice.deleteDocument(id);
        toast.success(`Deleted ${name}`);
      } catch (e) {
        toast.error(`Failed to delete ${name}`);
      }
    }
    setSelectedIds([]);
    onUpdate();
  }

  const onRenameOpen = () => {
    if (selectedIds.length !== 1) return;
    const file = folder.children.find(f => f.id === selectedIds[0]);
    setNewName(file?.data?.name || "");
    setShowRename(true);
  }

  const onRenameClick = async () => {
    const name = newName.trim();
    if (name === "") return;
    const id = selectedIds[0];
    try {
      await apiservice.renameDocument(id, name);
      toast.success(`Renamed to ${name}`);
    } catch (e) {
      toast.error("Failed to rename");
    }
    setShowRename(false);
    setSelectedIds([]);
    onUpdate();
  }

  const onMoveOpen = () => {
    if (selectedIds.length === 0) return;
    setTargetId("");
    setShowMove(true);
  }

  const onMoveClick = async () => {
    for (const id of selectedIds) {
      const file = folder.children.find(f => f.id === id);
      const name = file?.data?.name || id;
      try {
        await apiservice.moveDocument(id, targetId);
        toast.success(`Moved ${name}`);
      } catch (e) {
        toast.error(`Failed to move ${name}`);
      }
    }
    setShowMove(false);
    setSelectedIds([]);
    onUpdate();
  }

  const fileUploaded = () => {
    onUpdate();
  }

  const handleSelectItem = (id) => {
    setSelectedIds((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]
    );
  };

  // this should generally not happen, but just in case
  if (!folder) {
    return "nothing selected"
  }
  return (
    <>
      <Navbar style={{ marginLeft: '-12px' }}>
        { folder && (<div><NameTag node={folder} onSelect={onSelect} /></div>) }
      </Navbar>

      <Navbar className={styles.filedivider}>
        <Button size="sm" variant="outline" onClick={() => setShowCreateFolder(true)}>Create Folder</Button>
        <div className={styles.stretch}></div>
        <Button size="sm" onClick={onRenameOpen} disabled={selectedIds.length !== 1}>Rename</Button>
        <Button size="sm" onClick={onMoveOpen} disabled={selectedIds.length === 0}>Move</Button>
        <Button size="sm" onClick={onDeleteClick} disabled={selectedIds.length === 0}>Delete</Button>
        <ToggleButtonGroup value={listStyle} onChange={(v) => setListStyle(v)} name="abc">
          <ToggleButton id="grid" name="grid" size="sm" value="grid" variant="outline">
            <BsFillGridFill />
          </ToggleButton>
          <ToggleButton id="list" name="list" size="sm" value="list" variant="outline">
            <FaList />
          </ToggleButton>
        </ToggleButtonGroup>
      </Navbar>

      <Upload filesUploaded={fileUploaded} uploadFolder={selection.id}></Upload>
      <FileList
        listStyle={listStyle}
        files={folder.children}
        onSelect={onSelect}
        selectedIds={selectedIds}
        onSelectItem={handleSelectItem}
      />

      <Modal show={showCreateFileModal} onHide={() => setShowCreateFolder(false)}>
        <Modal.Header closeButton>
          Create a new folder
        </Modal.Header>

        <Modal.Body>
          <InputGroup className="mb-3">
            <Form.Control autoFocus={true} type="text" value={folderName} onChange={(e) => setFolderName(e.currentTarget.value)} />

            <Button variant="primary" onClick={onCreateFolderClick}>Create</Button>

          </InputGroup>
        </Modal.Body>
      </Modal>

      <Modal show={showRenameModal} onHide={() => setShowRename(false)}>
        <Modal.Header closeButton>
          Rename
        </Modal.Header>

        <Modal.Body>
          <InputGroup className="mb-3">
            <Form.Control autoFocus={true} type="text" value={newName}
              onChange={(e) => setNewName(e.currentTarget.value)}
              onKeyDown={(e) => { if (e.key === "Enter") onRenameClick(); }} />

            <Button variant="primary" onClick={onRenameClick} disabled={newName.trim() === ""}>Rename</Button>

          </InputGroup>
        </Modal.Body>
      </Modal>

      <Modal show={showMoveModal} onHide={() => setShowMove(false)}>
        <Modal.Header closeButton>
          Move to
        </Modal.Header>

        <Modal.Body>
          <InputGroup className="mb-3">
            <Form.Select value={targetId} onChange={(e) => setTargetId(e.currentTarget.value)}>
              <option value="">My Files</option>
              {folderOptions(entries?.find(e => e.id === "root")?.children, selectedIds).map(f => (
                <option key={f.id} value={f.id}>{"\u00a0\u00a0".repeat(f.depth)}{f.name}</option>
              ))}
            </Form.Select>

            <Button variant="primary" onClick={onMoveClick}>Move</Button>
          </InputGroup>
        </Modal.Body>
      </Modal>
    </>
  );
}
