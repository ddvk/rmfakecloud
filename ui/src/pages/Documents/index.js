import Tree from "./Tree";
import { useState } from "react";
import { Container, Row, Col } from "react-bootstrap";
import constants from "../../common/constants";
import File from "./File";
import Folder from "./Folder";
import './Documents.scss';
import Navbar from 'react-bootstrap/Navbar';

export default function DocumentList() {
  const [folder, setFolder] = useState(null);
  const [file, setFile] = useState(null);

  const onNodeSelected = (node) => {
    if (node && node.isFolder) {
      setFolder(node);
      setFile(null);
    } else {
      if (node) {
        node.downloadUrl = `${constants.ROOT_URL}/documents/${node.id}`;
      }
      setFile(node);
    }
  };

  return (
    <Container>
      <Row className="mt-2">
        <Col md={4}>
          <Navbar>
            <h6>My Documents</h6>
          </Navbar>
          <Tree onFileSelected={onNodeSelected} />
        </Col>
        <Col md={8}>
          {file && (<File file={file} onClose={() => setFile(null)} />)}
          {!file && folder && <Folder folder={folder} />}
        </Col>
      </Row>
    </Container>
  );
}
