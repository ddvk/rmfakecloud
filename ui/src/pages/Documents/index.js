import Tree from "./Tree";
import { useState } from "react";
import { Container, Row, Col } from "react-bootstrap";
import constants from "../../common/constants";
import File from "./File";
import './Documents.css';

export default function DocumentList() {
  //const [folder, setFolder] = useState("");
  const [file, setFile] = useState(null);

  const onFileSelected = (node) => {
    node.downloadUrl = `${constants.ROOT_URL}/documents/${node.id}`;
    setFile(node);
  };

  return (
    <Container>
      <Row>
        <Col md={4} className="mt-4">
          <h6>My Documents</h6>
          <Tree onFileSelected={onFileSelected} />
        </Col>
        <Col md={8}>
          {file && (<File file={file} />)}
        </Col>
      </Row>
    </Container>
  );
}
