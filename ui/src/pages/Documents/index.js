import Tree from "./Tree";
import { useState } from "react";
import { Container, Row, Col } from "react-bootstrap";
import constants from "../../common/constants";
import File from "./File";
import Folder from "./Folder";
import './Documents.scss';
import Navbar from 'react-bootstrap/Navbar';
import { BsSearch } from "react-icons/bs";
import Form from 'react-bootstrap/Form';
import Button from 'react-bootstrap/Button';
import InputGroup from 'react-bootstrap/InputGroup';

export default function DocumentList() {
  const [folder, setFolder] = useState(null);
  const [file, setFile] = useState(null);

  const [term, setTerm] = useState("");
  const [showSearch, setShowSearch] = useState(false);

  const onNodeSelected = (node) => {
    const file = node.data;

    if (file && file.isFolder) {
      setFolder(file);
      setFile(null);
    } else if (file) {
      file.downloadUrl = `${constants.ROOT_URL}/documents/${file.id}`;
      setFile(file);
    }
  };

  return (
    <Container>
      <Row className="mt-2">
        <Col md={4}>
          <Navbar>
            <h6 style={{ flex: 1}}>My Documents</h6>
            <h6>
              <Button variant="outline" onClick={() => { setShowSearch(!showSearch); setTerm("") }}><BsSearch/></Button>
            </h6>
          </Navbar>

          {showSearch && <div>
            <InputGroup className="mb-3">
              <InputGroup.Text>
                <BsSearch />
              </InputGroup.Text>

              <Form.Control autoFocus size="sm" type="text" value={term} onChange={(e) => { setTerm(e.currentTarget.value); }} />
            </InputGroup>
          </div>}

          <Tree onNodeSelected={onNodeSelected} term={term} />
        </Col>
        <Col md={8}>
          {file && (<File file={file} onClose={() => setFile(null)} />)}
          {!file && folder && <Folder folder={folder} />}
        </Col>
      </Row>
    </Container>
  );
}
