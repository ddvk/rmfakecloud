import Tree from "./Tree";
import { useState } from "react";
import { Container, Row, Col } from "react-bootstrap";
import File from "./File";
import Folder from "./Folder";
import './Documents.scss';
import Navbar from 'react-bootstrap/Navbar';
import { BsSearch } from "react-icons/bs";
import Form from 'react-bootstrap/Form';
import Button from 'react-bootstrap/Button';
import InputGroup from 'react-bootstrap/InputGroup';

export default function DocumentList() {
  const [selectedId, setSelectedId] = useState("root");
  const [selected, setSelected] = useState(null);
  const [term, setTerm] = useState("");
  const [showSearch, setShowSearch] = useState(false);

  const isFolder = selected && selected.data && selected.data.isFolder;

  const onSelect = (node) => {
    setSelected(node);
  };

  const onSelectById = (id) => {
    setSelectedId(id);
  };

  return (
    <Container>
      <Row className="mt-2">
        <Col md={4}>
          <Navbar>
            <div style={{ flex: 1, fontWeight: 'bold' }}>My Documents</div>
            <Button variant="outline" onClick={() => { setShowSearch(!showSearch); setTerm("") }}><BsSearch/></Button>
          </Navbar>

          {showSearch && <div>
            <InputGroup className="mb-3">
              <InputGroup.Text>
                <BsSearch />
              </InputGroup.Text>

              <Form.Control autoFocus size="sm" type="text" value={term} onChange={(e) => { setTerm(e.currentTarget.value); }} />
            </InputGroup>
          </div>}

          <Tree selection={selectedId} onSelect={onSelect} term={term} />
        </Col>
        <Col md={8}>
          {selected && !isFolder && <File file={selected} onSelect={onSelectById} />}
          {selected && isFolder && <Folder folder={selected} onSelect={onSelectById} />}
        </Col>
      </Row>
    </Container>
  );
}
